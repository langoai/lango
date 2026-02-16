package agent

import (
	"context"
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

// Tool represents a tool that can be invoked by the LLM
type Tool struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
	Handler     ToolHandler
	SafetyLevel SafetyLevel
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
