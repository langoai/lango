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
	t.sender(ApprovalRequestMsg{Request: req, Response: respCh})

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

// renderApprovalBanner renders the inline approval prompt.
func renderApprovalBanner(req approval.ApprovalRequest, width int) string {
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(tui.Warning).
		Width(width - 4).
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

	keys := tui.HelpBar(
		tui.HelpEntry("a", "allow"),
		tui.HelpEntry("s", "allow session"),
		tui.HelpEntry("d", "deny"),
		tui.HelpEntry("esc", "deny"),
	)

	content := fmt.Sprintf("%s\n%s  %s%s\n\n%s", title, tool, summary, params, keys)
	return border.Render(content)
}
