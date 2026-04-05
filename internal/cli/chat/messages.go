package chat

import (
	"time"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/turnrunner"
)

// ChunkMsg delivers a streaming text chunk from the agent.
type ChunkMsg struct {
	Chunk string
}

// DoneMsg signals that a turn has finished.
type DoneMsg struct {
	Result turnrunner.Result
}

// ErrorMsg signals a runtime error.
type ErrorMsg struct {
	Err error
}

// WarningMsg signals that a turn is approaching its timeout.
type WarningMsg struct {
	Elapsed     time.Duration
	HardCeiling time.Duration
}

// ApprovalRequestMsg delivers an approval request from the agent runtime.
type ApprovalRequestMsg struct {
	Request   approval.ApprovalRequest
	ViewModel approval.ApprovalViewModel
	Response  chan<- approval.ApprovalResponse
}

// SystemMsg is a locally generated informational message.
type SystemMsg struct {
	Text string
}

// CursorTickMsg triggers cursor blink toggle during streaming.
type CursorTickMsg time.Time

// ToolStartedMsg signals that a tool invocation has begun.
type ToolStartedMsg struct {
	CallID   string
	ToolName string
	Params   map[string]any
}

// ToolFinishedMsg signals that a tool invocation has completed.
type ToolFinishedMsg struct {
	CallID   string
	ToolName string
	Success  bool
	Duration time.Duration
	Output   string
}

// ThinkingStartedMsg signals that the agent has started thinking/reasoning.
type ThinkingStartedMsg struct {
	AgentName string
	Summary   string
}

// ThinkingFinishedMsg signals that the agent has finished thinking/reasoning.
type ThinkingFinishedMsg struct {
	AgentName string
	Duration  time.Duration
	Summary   string
}

// TaskStripTickMsg triggers periodic task strip refresh.
type TaskStripTickMsg time.Time

// PendingIndicatorTickMsg triggers pending indicator update (submit → first event).
type PendingIndicatorTickMsg time.Time

// ChannelMessageMsg is sent when a channel message is received via EventBus.
type ChannelMessageMsg struct {
	Channel    string
	SessionKey string
	SenderName string
	Text       string
	Timestamp  time.Time
	Metadata   map[string]string
}
