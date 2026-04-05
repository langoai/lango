package chat

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/cli/tui"
)

// TUIApprovalProvider implements approval.Provider for the interactive TUI.
// It sends approval requests to the bubbletea program and waits for the
// user to respond with a/s/d keys.
type TUIApprovalProvider struct {
	sender func(msg interface{})
}

// NewTUIApprovalProvider creates a TUI approval provider that uses the given
// bubbletea program send function for dispatching approval UI messages.
func NewTUIApprovalProvider(sender func(msg interface{})) *TUIApprovalProvider {
	return &TUIApprovalProvider{sender: sender}
}

// RequestApproval sends an approval request to the TUI and blocks until
// the user responds or the context is cancelled.
func (t *TUIApprovalProvider) RequestApproval(ctx context.Context, req approval.ApprovalRequest) (approval.ApprovalResponse, error) {
	respCh := make(chan approval.ApprovalResponse, 1)
	vm := approval.NewViewModel(req)
	t.sender(ApprovalRequestMsg{Request: req, ViewModel: vm, Response: respCh})

	select {
	case resp := <-respCh:
		if resp.Provider == "" {
			resp.Provider = "tui"
		}
		return resp, nil
	case <-ctx.Done():
		return approval.ApprovalResponse{}, ctx.Err()
	}
}

// CanHandle returns false — TUI provider is a fallback, not a session-prefix router.
func (t *TUIApprovalProvider) CanHandle(_ string) bool { return false }

// Name returns the provider name for logging.
func (t *TUIApprovalProvider) Name() string { return "tui" }

// renderApproval dispatches to the appropriate approval renderer based on tier.
func renderApproval(msg *ApprovalRequestMsg, width, height int) string {
	switch msg.ViewModel.Tier {
	case approval.TierFullscreen:
		return renderApprovalDialog(msg.ViewModel, width, height)
	case approval.TierInline:
		return renderApprovalStrip(msg.ViewModel, width)
	default:
		return renderApprovalBanner(msg.Request, width)
	}
}

// formatChannelOrigin extracts channel origin info from a session key.
// Returns a human-readable string like "[Telegram] 123456" or "" for non-channel keys.
func formatChannelOrigin(sessionKey string) string {
	parts := strings.SplitN(sessionKey, ":", 3)
	if len(parts) < 3 {
		return ""
	}

	var channelName string
	switch parts[0] {
	case "telegram":
		channelName = "Telegram"
	case "discord":
		channelName = "Discord"
	case "slack":
		channelName = "Slack"
	default:
		return ""
	}

	return fmt.Sprintf("[%s] %s", channelName, parts[1])
}

// formatChannelBadge returns a short channel badge for compact displays.
// Returns "" for non-channel session keys.
func formatChannelBadge(sessionKey string) string {
	parts := strings.SplitN(sessionKey, ":", 3)
	if len(parts) < 3 {
		return ""
	}

	switch parts[0] {
	case "telegram":
		return "[TG]"
	case "discord":
		return "[DC]"
	case "slack":
		return "[SL]"
	default:
		return ""
	}
}

// renderApprovalBanner renders the inline approval prompt.
func renderApprovalBanner(req approval.ApprovalRequest, width int) string {
	bannerWidth := width - 4
	if bannerWidth < 10 {
		bannerWidth = 10
	}
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.Warning).
		Width(bannerWidth).
		Padding(0, 1)

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.Warning).
		Render("Tool Approval Required")

	tool := lipgloss.NewStyle().
		Bold(true).
		Foreground(tui.Highlight).
		Render(req.ToolName)

	summary := req.Summary
	if summary == "" {
		summary = fmt.Sprintf("Execute tool: %s", req.ToolName)
	}

	var params string
	if len(req.Params) > 0 {
		var parts []string
		for k, v := range req.Params {
			parts = append(parts, fmt.Sprintf("  %s: %v", k, v))
		}
		params = "\n" + lipgloss.NewStyle().Foreground(tui.Muted).Render(strings.Join(parts, "\n"))
	}

	var originLine string
	if origin := formatChannelOrigin(req.SessionKey); origin != "" {
		originLine = "\n" + lipgloss.NewStyle().Foreground(tui.Info).Render("  ← "+origin)
	}

	keys := tui.HelpBar(
		tui.HelpEntry("a", "allow"),
		tui.HelpEntry("s", "allow session"),
		tui.HelpEntry("d", "deny"),
		tui.HelpEntry("esc", "deny"),
	)

	content := fmt.Sprintf("%s\n%s  %s%s%s\n\n%s", title, tool, summary, originLine, params, keys)
	return border.Render(content)
}
