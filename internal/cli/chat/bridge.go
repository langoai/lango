package chat

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/langoai/lango/internal/turnrunner"
)

// enrichRequest wires turnrunner callbacks to Bubble Tea messages
// without overwriting existing OnChunk/OnWarning callbacks.
func enrichRequest(program *tea.Program, req *turnrunner.Request) {
	if program == nil {
		return
	}

	req.OnToolCall = func(callID, toolName string, params map[string]any) {
		program.Send(ToolStartedMsg{
			CallID:   callID,
			ToolName: toolName,
			Params:   params,
		})
	}

	req.OnToolResult = func(callID, toolName string, success bool, duration time.Duration, preview string) {
		program.Send(ToolFinishedMsg{
			CallID:   callID,
			ToolName: toolName,
			Success:  success,
			Duration: duration,
			Output:   preview,
		})
	}

	req.OnThinking = func(agentName string, started bool, summary string) {
		if started {
			program.Send(ThinkingStartedMsg{
				AgentName: agentName,
				Summary:   summary,
			})
		} else {
			program.Send(ThinkingFinishedMsg{
				AgentName: agentName,
			})
		}
	}
}
