package agent

import (
	"context"
	"iter"
)

// SafetyLevel classifies the risk level of a tool.
// Zero value is treated as Dangerous (fail-safe).
type SafetyLevel int

const (
	// SafetyLevelSafe indicates a read-only or non-destructive tool.
	SafetyLevelSafe SafetyLevel = iota + 1
	// SafetyLevelModerate indicates a tool that creates or modifies non-critical resources.
	SafetyLevelModerate
	// SafetyLevelDangerous indicates a tool that can execute arbitrary code, delete data, or modify secrets.
	SafetyLevelDangerous
)

// Valid reports whether s is a known safety level.
func (s SafetyLevel) Valid() bool {
	switch s {
	case SafetyLevelSafe, SafetyLevelModerate, SafetyLevelDangerous:
		return true
	}
	return false
}

// Values returns all known safety levels.
func (s SafetyLevel) Values() []SafetyLevel {
	return []SafetyLevel{SafetyLevelSafe, SafetyLevelModerate, SafetyLevelDangerous}
}

// String returns the human-readable name of the safety level.
func (s SafetyLevel) String() string {
	switch s {
	case SafetyLevelSafe:
		return "safe"
	case SafetyLevelModerate:
		return "moderate"
	case SafetyLevelDangerous:
		return "dangerous"
	default:
		return "dangerous" // zero value → fail-safe
	}
}

// IsDangerous returns true if the tool should be treated as dangerous.
// Zero value (unset) is also treated as dangerous.
func (s SafetyLevel) IsDangerous() bool {
	return s == SafetyLevelDangerous || s == 0
}

// ParseSafetyLevel converts a string to a SafetyLevel.
// Returns SafetyLevelDangerous and false for unknown strings (fail-safe).
func ParseSafetyLevel(s string) (SafetyLevel, bool) {
	switch s {
	case "safe":
		return SafetyLevelSafe, true
	case "moderate":
		return SafetyLevelModerate, true
	case "dangerous":
		return SafetyLevelDangerous, true
	default:
		return SafetyLevelDangerous, false
	}
}

// Tool represents a tool that can be invoked by the LLM
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
	Handler     ToolHandler
	// StreamingHandler is an optional streaming implementation. When non-nil,
	// AdaptStreamingTool routes the tool through ADK's functiontool.NewStreaming.
	// Handler is still required as the non-streaming fallback.
	StreamingHandler StreamingToolHandler
	SafetyLevel      SafetyLevel
	Capability       ToolCapability // Zero value = backward compatible defaults
}

// ParameterDef defines a tool parameter
type ParameterDef struct {
	Type        string
	Description string
	Required    bool
	Enum        []string
}

// ToolHandler is the function signature for tool implementations
type ToolHandler func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// StreamingToolHandler is the function signature for streaming tool implementations.
// Yields zero or more partial result strings followed by an optional error.
// In non-live agent runs, the runtime aggregates all yielded strings into a single
// final result before returning to the model. In Live API runs (Track E), each
// yielded chunk is delivered to the live session as it is produced.
type StreamingToolHandler func(ctx context.Context, params map[string]interface{}) iter.Seq2[string, error]
