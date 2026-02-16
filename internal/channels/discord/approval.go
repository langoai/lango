package discord

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/langowarny/lango/internal/approval"
)

// ApprovalProvider implements approval.Provider for Discord using Button components.
type ApprovalProvider struct {
	session Session
	pending sync.Map // map[requestID]chan bool
	timeout time.Duration
}

var _ approval.Provider = (*ApprovalProvider)(nil)

// NewApprovalProvider creates a Discord approval provider.
func NewApprovalProvider(sess Session, timeout time.Duration) *ApprovalProvider {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &ApprovalProvider{
		session: sess,
		timeout: timeout,
	}
}

// RequestApproval sends a message with approve/deny buttons and waits for interaction.
func (p *ApprovalProvider) RequestApproval(ctx context.Context, req approval.ApprovalRequest) (bool, error) {
	channelID, err := parseDiscordChannelID(req.SessionKey)
	if err != nil {
		return false, fmt.Errorf("parse session key: %w", err)
	}

	respChan := make(chan bool, 1)
	p.pending.Store(req.ID, respChan)
	defer p.pending.Delete(req.ID)

	content := fmt.Sprintf("🔐 Tool **%s** requires approval", req.ToolName)
	if req.Summary != "" {
		content += "\n```\n" + req.Summary + "\n```"
	}
	_, err = p.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: content,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Approve",
						Style:    discordgo.SuccessButton,
						CustomID: "approve:" + req.ID,
						Emoji: &discordgo.ComponentEmoji{
							Name: "✅",
						},
					},
					discordgo.Button{
						Label:    "Deny",
						Style:    discordgo.DangerButton,
						CustomID: "deny:" + req.ID,
						Emoji: &discordgo.ComponentEmoji{
							Name: "❌",
						},
					},
				},
			},
		},
	})
	if err != nil {
		return false, fmt.Errorf("send approval message: %w", err)
	}

	select {
	case approved := <-respChan:
		return approved, nil
	case <-ctx.Done():
		return false, ctx.Err()
	case <-time.After(p.timeout):
		return false, fmt.Errorf("approval timeout")
	}
}

// HandleInteraction processes a button interaction for approval.
func (p *ApprovalProvider) HandleInteraction(i *discordgo.InteractionCreate) {
	if i == nil || i.Type != discordgo.InteractionMessageComponent {
		return
	}

	data := i.MessageComponentData()
	customID := data.CustomID

	var requestID string
	var approved bool

	if strings.HasPrefix(customID, "approve:") {
		requestID = strings.TrimPrefix(customID, "approve:")
		approved = true
	} else if strings.HasPrefix(customID, "deny:") {
		requestID = strings.TrimPrefix(customID, "deny:")
		approved = false
	} else {
		return
	}

	// Respond to the interaction
	status := "❌ Denied"
	if approved {
		status = "✅ Approved"
	}

	err := p.session.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    fmt.Sprintf("🔐 Tool approval — %s", status),
			Components: []discordgo.MessageComponent{}, // remove buttons
		},
	})
	if err != nil {
		logger.Warnw("interaction respond error", "error", err)
	}

	// Send result to waiting goroutine
	if ch, ok := p.pending.LoadAndDelete(requestID); ok {
		respChan := ch.(chan bool)
		select {
		case respChan <- approved:
		default:
		}
	}
}

// CanHandle returns true for session keys starting with "discord:".
func (p *ApprovalProvider) CanHandle(sessionKey string) bool {
	return strings.HasPrefix(sessionKey, "discord:")
}

// parseDiscordChannelID extracts the channelID from a session key like "discord:<channelID>:<userID>".
func parseDiscordChannelID(sessionKey string) (string, error) {
	parts := strings.SplitN(sessionKey, ":", 3)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid discord session key: %s", sessionKey)
	}
	return parts[1], nil
}
