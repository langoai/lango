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

// approvalPending holds the response channel and message metadata for a pending approval.
type approvalPending struct {
	ch        chan approval.ApprovalResponse
	channelID string
	messageID string
}

// ApprovalProvider implements approval.Provider for Discord using Button components.
type ApprovalProvider struct {
	session Session
	pending sync.Map // map[requestID]*approvalPending
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

// RequestApproval sends a message with approve/deny/always-allow buttons and waits for interaction.
func (p *ApprovalProvider) RequestApproval(ctx context.Context, req approval.ApprovalRequest) (approval.ApprovalResponse, error) {
	channelID, err := parseDiscordChannelID(req.SessionKey)
	if err != nil {
		return approval.ApprovalResponse{}, fmt.Errorf("parse session key: %w", err)
	}

	respChan := make(chan approval.ApprovalResponse, 1)

	content := fmt.Sprintf("🔐 Tool **%s** requires approval", req.ToolName)
	if req.Summary != "" {
		content += "\n```\n" + req.Summary + "\n```"
	}
	sentMsg, err := p.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
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
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Always Allow",
						Style:    discordgo.SecondaryButton,
						CustomID: "always:" + req.ID,
						Emoji: &discordgo.ComponentEmoji{
							Name: "🔓",
						},
					},
				},
			},
		},
	})
	if err != nil {
		return approval.ApprovalResponse{}, fmt.Errorf("send approval message: %w", err)
	}

	p.pending.Store(req.ID, &approvalPending{
		ch:        respChan,
		channelID: channelID,
		messageID: sentMsg.ID,
	})
	defer p.pending.Delete(req.ID)

	select {
	case resp := <-respChan:
		return resp, nil
	case <-ctx.Done():
		p.editExpiredMessage(channelID, sentMsg.ID)
		return approval.ApprovalResponse{}, ctx.Err()
	case <-time.After(p.timeout):
		p.editExpiredMessage(channelID, sentMsg.ID)
		return approval.ApprovalResponse{}, fmt.Errorf("approval timeout")
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
	var resp approval.ApprovalResponse

	if strings.HasPrefix(customID, "approve:") {
		requestID = strings.TrimPrefix(customID, "approve:")
		resp = approval.ApprovalResponse{Approved: true}
	} else if strings.HasPrefix(customID, "deny:") {
		requestID = strings.TrimPrefix(customID, "deny:")
		resp = approval.ApprovalResponse{}
	} else if strings.HasPrefix(customID, "always:") {
		requestID = strings.TrimPrefix(customID, "always:")
		resp = approval.ApprovalResponse{Approved: true, AlwaysAllow: true}
	} else {
		return
	}

	// Respond to the interaction
	var status string
	switch {
	case resp.AlwaysAllow:
		status = "🔓 Always Allowed"
	case resp.Approved:
		status = "✅ Approved"
	default:
		status = "❌ Denied"
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
	if val, ok := p.pending.LoadAndDelete(requestID); ok {
		pending, ok := val.(*approvalPending)
		if !ok {
			logger.Warnw("unexpected pending type", "requestId", requestID)
			return
		}
		select {
		case pending.ch <- resp:
		default:
		}
	}
}

// CanHandle returns true for session keys starting with "discord:".
func (p *ApprovalProvider) CanHandle(sessionKey string) bool {
	return strings.HasPrefix(sessionKey, "discord:")
}

// editExpiredMessage updates the approval message to show expired status and removes buttons.
func (p *ApprovalProvider) editExpiredMessage(channelID, messageID string) {
	content := "🔐 Tool approval — ⏱ Expired"
	emptyComponents := []discordgo.MessageComponent{}
	_, err := p.session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    channelID,
		ID:         messageID,
		Content:    &content,
		Components: &emptyComponents,
	})
	if err != nil {
		logger.Warnw("edit expired approval message error", "error", err)
	}
}

// parseDiscordChannelID extracts the channelID from a session key like "discord:<channelID>:<userID>".
func parseDiscordChannelID(sessionKey string) (string, error) {
	parts := strings.SplitN(sessionKey, ":", 3)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid discord session key: %s", sessionKey)
	}
	return parts[1], nil
}
