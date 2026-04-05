package chat

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/langoai/lango/internal/turnrunner"
)

// msgSender abstracts tea.Program.Send for testability.
type msgSender interface {
	Send(msg tea.Msg)
}

// enrichRequest wires turnrunner callbacks to Bubble Tea messages
// without overwriting existing OnChunk/OnWarning callbacks.
func enrichRequest(sender msgSender, req *turnrunner.Request) {
	if sender == nil {
		return
	}

	req.OnToolCall = func(callID, toolName string, params map[string]any) {
		sender.Send(ToolStartedMsg{
			CallID:   callID,
			ToolName: toolName,
			Params:   params,
		})
	}

	req.OnToolResult = func(callID, toolName string, success bool, duration time.Duration, preview string) {
		sender.Send(ToolFinishedMsg{
			CallID:   callID,
			ToolName: toolName,
			Success:  success,
			Duration: duration,
			Output:   preview,
		})
	}

	req.OnDelegation = func(from, to, reason string) {
		sender.Send(DelegationMsg{From: from, To: to, Reason: reason})
	}

	req.OnBudgetWarning = func(used, max int) {
		sender.Send(BudgetWarningMsg{Used: used, Max: max})
	}

	var thinkingStart time.Time
	req.OnThinking = func(agentName string, started bool, summary string) {
		if started {
			thinkingStart = time.Now()
			sender.Send(ThinkingStartedMsg{
				AgentName: agentName,
				Summary:   summary,
			})
		} else {
			sender.Send(ThinkingFinishedMsg{
				AgentName: agentName,
				Duration:  time.Since(thinkingStart),
				Summary:   summary,
			})
		}
	}
}
