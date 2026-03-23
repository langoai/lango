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
	Request  approval.ApprovalRequest
	Response chan<- approval.ApprovalResponse
}

// SystemMsg is a locally generated informational message.
type SystemMsg struct {
	Text string
}
