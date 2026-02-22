package slack

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/langoai/lango/internal/approval"
	slackapi "github.com/slack-go/slack"
)

// approvalPending holds the response channel and message metadata for a pending approval.
type approvalPending struct {
	ch        chan approval.ApprovalResponse
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

// RequestApproval posts a message with approve/deny/always-allow action buttons and waits for interaction.
func (p *ApprovalProvider) RequestApproval(ctx context.Context, req approval.ApprovalRequest) (approval.ApprovalResponse, error) {
	channelID, err := parseSlackChannelID(req.SessionKey)
	if err != nil {
		return approval.ApprovalResponse{}, fmt.Errorf("parse session key: %w", err)
	}

	respChan := make(chan approval.ApprovalResponse, 1)

	approveBtn := slackapi.NewButtonBlockElement("approve:"+req.ID, "approve",
		slackapi.NewTextBlockObject("plain_text", "✅ Approve", true, false))
	approveBtn.Style = slackapi.StylePrimary

	denyBtn := slackapi.NewButtonBlockElement("deny:"+req.ID, "deny",
		slackapi.NewTextBlockObject("plain_text", "❌ Deny", true, false))
	denyBtn.Style = slackapi.StyleDanger

	alwaysBtn := slackapi.NewButtonBlockElement("always:"+req.ID, "always",
		slackapi.NewTextBlockObject("plain_text", "🔓 Always Allow", true, false))

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
				alwaysBtn,
			),
		),
	)
	if err != nil {
		return approval.ApprovalResponse{}, fmt.Errorf("send approval message: %w", err)
	}

	p.pending.Store(req.ID, &approvalPending{
		ch:        respChan,
		channelID: channelID,
		timestamp: ts,
	})
	defer p.pending.Delete(req.ID)

	select {
	case resp := <-respChan:
		return resp, nil
	case <-ctx.Done():
		p.editExpiredMessage(channelID, ts)
		return approval.ApprovalResponse{}, ctx.Err()
	case <-time.After(p.timeout):
		p.editExpiredMessage(channelID, ts)
		return approval.ApprovalResponse{}, fmt.Errorf("approval timeout")
	}
}

// HandleInteractive processes a Slack interactive callback (block_actions) for approval.
func (p *ApprovalProvider) HandleInteractive(actionID string) {
	var requestID string
	var resp approval.ApprovalResponse

	if strings.HasPrefix(actionID, "approve:") {
		requestID = strings.TrimPrefix(actionID, "approve:")
		resp = approval.ApprovalResponse{Approved: true}
	} else if strings.HasPrefix(actionID, "deny:") {
		requestID = strings.TrimPrefix(actionID, "deny:")
		resp = approval.ApprovalResponse{}
	} else if strings.HasPrefix(actionID, "always:") {
		requestID = strings.TrimPrefix(actionID, "always:")
		resp = approval.ApprovalResponse{Approved: true, AlwaysAllow: true}
	} else {
		return
	}

	// LoadAndDelete first to prevent TOCTOU race
	val, ok := p.pending.LoadAndDelete(requestID)
	if !ok {
		return // already processed or expired
	}

	pending, ok := val.(*approvalPending)
	if !ok {
		logger.Warnw("unexpected pending type", "requestId", requestID)
		return
	}

	// Update the original message to remove buttons
	var status string
	switch {
	case resp.AlwaysAllow:
		status = "🔓 Always Allowed"
	case resp.Approved:
		status = "✅ Approved"
	default:
		status = "❌ Denied"
	}
	_, _, _, err := p.api.UpdateMessage(pending.channelID, pending.timestamp,
		slackapi.MsgOptionText(fmt.Sprintf("🔐 Tool approval — %s", status), false),
		slackapi.MsgOptionBlocks(), // empty blocks = remove action buttons
	)
	if err != nil {
		logger.Warnw("update approval message error", "error", err)
	}

	// Send result to waiting goroutine
	select {
	case pending.ch <- resp:
	default:
	}
}

// CanHandle returns true for session keys starting with "slack:".
func (p *ApprovalProvider) CanHandle(sessionKey string) bool {
	return strings.HasPrefix(sessionKey, "slack:")
}

// editExpiredMessage updates the approval message to show expired status and removes buttons.
func (p *ApprovalProvider) editExpiredMessage(channelID, timestamp string) {
	_, _, _, err := p.api.UpdateMessage(channelID, timestamp,
		slackapi.MsgOptionText("🔐 Tool approval — ⏱ Expired", false),
		slackapi.MsgOptionBlocks(), // remove action buttons
	)
	if err != nil {
		logger.Warnw("edit expired approval message error", "error", err)
	}
}

// parseSlackChannelID extracts the channelID from a session key like "slack:<channelID>:<userID>".
func parseSlackChannelID(sessionKey string) (string, error) {
	parts := strings.SplitN(sessionKey, ":", 3)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid slack session key: %s", sessionKey)
	}
	return parts[1], nil
}
