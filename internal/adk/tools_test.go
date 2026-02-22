package adk

import (
	"context"
	"testing"

	"github.com/langoai/lango/internal/agent"
)

func TestAdaptTool_ParameterDef(t *testing.T) {
	tool := &agent.Tool{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters: map[string]interface{}{
			"command": agent.ParameterDef{
				Type:        "string",
				Description: "The command to run",
				Required:    true,
			},
			"timeout": agent.ParameterDef{
				Type:        "integer",
				Description: "Timeout in seconds",
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "ok", nil
		},
	}

	adkTool, err := AdaptTool(tool)
	if err != nil {
		t.Fatalf("AdaptTool failed: %v", err)
	}
	if adkTool == nil {
		t.Fatal("expected non-nil tool")
	}
}

func TestAdaptTool_MapParams(t *testing.T) {
	tool := &agent.Tool{
		Name:        "map_tool",
		Description: "A tool with map params",
		Parameters: map[string]interface{}{
			"arg": map[string]interface{}{
				"type":        "string",
				"description": "An argument",
				"required":    true,
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "result", nil
		},
	}

	adkTool, err := AdaptTool(tool)
	if err != nil {
		t.Fatalf("AdaptTool failed: %v", err)
	}
	if adkTool == nil {
		t.Fatal("expected non-nil tool")
	}
}

func TestAdaptTool_FallbackParams(t *testing.T) {
	// Test with an unknown param type (not ParameterDef, not map)
	tool := &agent.Tool{
		Name:        "fallback_tool",
		Description: "A tool with fallback params",
		Parameters: map[string]interface{}{
			"arg": "just a string", // Neither ParameterDef nor map
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	}

	adkTool, err := AdaptTool(tool)
	if err != nil {
		t.Fatalf("AdaptTool failed: %v", err)
	}
	if adkTool == nil {
		t.Fatal("expected non-nil tool")
	}
}

func TestAdaptTool_NoParams(t *testing.T) {
	tool := &agent.Tool{
		Name:        "no_params_tool",
		Description: "A tool with no params",
		Parameters:  map[string]interface{}{},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return "done", nil
		},
	}

	adkTool, err := AdaptTool(tool)
	if err != nil {
		t.Fatalf("AdaptTool failed: %v", err)
	}
	if adkTool == nil {
		t.Fatal("expected non-nil tool")
	}
}

func TestAdaptTool_WithEnum(t *testing.T) {
	tool := &agent.Tool{
		Name:        "enum_tool",
		Description: "A tool with enum param",
		Parameters: map[string]interface{}{
			"action": agent.ParameterDef{
				Type:        "string",
				Description: "Action to take",
				Required:    true,
				Enum:        []string{"start", "stop", "restart"},
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return params["action"], nil
		},
	}

	adkTool, err := AdaptTool(tool)
	if err != nil {
		t.Fatalf("AdaptTool failed: %v", err)
	}
	if adkTool == nil {
		t.Fatal("expected non-nil tool")
	}
}
