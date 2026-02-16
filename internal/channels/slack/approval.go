package slack

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	slackapi "github.com/slack-go/slack"
	"github.com/langowarny/lango/internal/approval"
)

// approvalPending holds the response channel and message metadata for a pending approval.
type approvalPending struct {
	ch        chan bool
	channelID string
	timestamp string
}

// ApprovalProvider implements approval.Provider for Slack using Block Kit action buttons.
type ApprovalProvider struct {
	api     Client
	pending sync.Map // map[requestID]*approvalPending
	timeout time.Duration
}

var _ approval.Provider = (*ApprovalProvider)(nil)

// NewApprovalProvider creates a Slack approval provider.
func NewApprovalProvider(api Client, timeout time.Duration) *ApprovalProvider {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &ApprovalProvider{
		api:     api,
		timeout: timeout,
	}
}

// RequestApproval posts a message with approve/deny action buttons and waits for interaction.
func (p *ApprovalProvider) RequestApproval(ctx context.Context, req approval.ApprovalRequest) (bool, error) {
	channelID, err := parseSlackChannelID(req.SessionKey)
	if err != nil {
		return false, fmt.Errorf("parse session key: %w", err)
	}

	respChan := make(chan bool, 1)

	approveBtn := slackapi.NewButtonBlockElement("approve:"+req.ID, "approve",
		slackapi.NewTextBlockObject("plain_text", "✅ Approve", true, false))
	approveBtn.Style = slackapi.StylePrimary

	denyBtn := slackapi.NewButtonBlockElement("deny:"+req.ID, "deny",
		slackapi.NewTextBlockObject("plain_text", "❌ Deny", true, false))
	denyBtn.Style = slackapi.StyleDanger

	sectionText := fmt.Sprintf("🔐 Tool *%s* requires approval", req.ToolName)
	if req.Summary != "" {
		sectionText += "\n```" + req.Summary + "```"
	}
	_, ts, err := p.api.PostMessage(channelID,
		slackapi.MsgOptionText(fmt.Sprintf("🔐 Tool '%s' requires approval", req.ToolName), false),
		slackapi.MsgOptionBlocks(
			slackapi.NewSectionBlock(
				slackapi.NewTextBlockObject("mrkdwn", sectionText, false, false),
				nil, nil,
			),
			slackapi.NewActionBlock(
				"approval_actions",
				approveBtn,
				denyBtn,
			),
		),
	)
	if err != nil {
		return false, fmt.Errorf("send approval message: %w", err)
	}

	p.pending.Store(req.ID, &approvalPending{
		ch:        respChan,
		channelID: channelID,
		timestamp: ts,
	})
	defer p.pending.Delete(req.ID)

	select {
	case approved := <-respChan:
		return approved, nil
	case <-ctx.Done():
		return false, ctx.Err()
	case <-time.After(p.timeout):
		return false, fmt.Errorf("approval timeout")
	}
}

// HandleInteractive processes a Slack interactive callback (block_actions) for approval.
func (p *ApprovalProvider) HandleInteractive(actionID string) {
	var requestID string
	var approved bool

	if strings.HasPrefix(actionID, "approve:") {
		requestID = strings.TrimPrefix(actionID, "approve:")
		approved = true
	} else if strings.HasPrefix(actionID, "deny:") {
		requestID = strings.TrimPrefix(actionID, "deny:")
		approved = false
	} else {
		return
	}

	// Update the original message to remove buttons
	if val, ok := p.pending.Load(requestID); ok {
		pending := val.(*approvalPending)
		status := "❌ Denied"
		if approved {
			status = "✅ Approved"
		}
		_, _, _, err := p.api.UpdateMessage(pending.channelID, pending.timestamp,
			slackapi.MsgOptionText(fmt.Sprintf("🔐 Tool approval — %s", status), false),
		)
		if err != nil {
			logger.Warnw("update approval message error", "error", err)
		}
	}

	// Send result to waiting goroutine
	if val, ok := p.pending.LoadAndDelete(requestID); ok {
		pending := val.(*approvalPending)
		select {
		case pending.ch <- approved:
		default:
		}
	}
}

// CanHandle returns true for session keys starting with "slack:".
func (p *ApprovalProvider) CanHandle(sessionKey string) bool {
	return strings.HasPrefix(sessionKey, "slack:")
}

// parseSlackChannelID extracts the channelID from a session key like "slack:<channelID>:<userID>".
func parseSlackChannelID(sessionKey string) (string, error) {
	parts := strings.SplitN(sessionKey, ":", 3)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid slack session key: %s", sessionKey)
	}
	return parts[1], nil
}
