// Package sandbox provides process isolation for tool execution.
// Remote peer tool invocations run in isolated subprocesses to prevent
// access to process memory (passphrases, private keys, session tokens).
package sandbox

import (
	"context"
	"time"
)

// Executor runs a tool invocation, optionally in an isolated subprocess.
type Executor interface {
	Execute(ctx context.Context, toolName string, params map[string]interface{}) (map[string]interface{}, error)
}

// Config controls sandbox execution behavior.
type Config struct {
	// Enabled turns on subprocess isolation for remote tool calls.
	Enabled bool

	// TimeoutPerTool is the maximum duration for a single tool execution.
	// Zero means no timeout.
	TimeoutPerTool time.Duration

	// MaxMemoryMB is a soft memory limit for the subprocess (Phase 2).
	MaxMemoryMB int
}

// ExecutionRequest is the JSON message sent to the sandbox worker via stdin.
type ExecutionRequest struct {
	ToolName string                 `json:"toolName"`
	Params   map[string]interface{} `json:"params"`
}

// ExecutionResult is the JSON message received from the sandbox worker via stdout.
type ExecutionResult struct {
	Output map[string]interface{} `json:"output,omitempty"`
	Error  string                 `json:"error,omitempty"`
}
