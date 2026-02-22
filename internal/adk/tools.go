package adk

import (
	"context"
	"fmt"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"github.com/langoai/lango/internal/agent"
)

// AdaptTool converts an internal agent.Tool to an ADK tool.Tool
func AdaptTool(t *agent.Tool) (tool.Tool, error) {
	// Build input schema from parameters
	props := make(map[string]*jsonschema.Schema)
	var required []string

	for name, paramDef := range t.Parameters {
		s := &jsonschema.Schema{}

		// Attempt to parse ParameterDef
		// Since it is stored as interface{}, we need to handle potential map conversions if it came from JSON
		// But in-memory tools usually use the struct.
		if pd, ok := paramDef.(agent.ParameterDef); ok {
			s.Type = pd.Type
			s.Description = pd.Description
			if len(pd.Enum) > 0 {
				s.Enum = make([]any, len(pd.Enum))
				for i, v := range pd.Enum {
					s.Enum[i] = v
				}
			}
			if pd.Required {
				required = append(required, name)
			}
		} else if pdMap, ok := paramDef.(map[string]interface{}); ok {
			// Handle map (e.g. from JSON config)
			if t, ok := pdMap["type"].(string); ok {
				s.Type = t
			}
			if d, ok := pdMap["description"].(string); ok {
				s.Description = d
			}
			if r, ok := pdMap["required"].(bool); ok && r {
				required = append(required, name)
			}
		} else {
			// Fallback or skip
			s.Type = "string" // default
		}
		props[name] = s
	}

	inputSchema := &jsonschema.Schema{
		Type:       "object",
		Properties: props,
		Required:   required,
	}

	cfg := functiontool.Config{
		Name:        t.Name,
		Description: t.Description,
		InputSchema: inputSchema,
	}

	// Wrapper handler
	handler := func(ctx tool.Context, args map[string]any) (any, error) {
		return t.Handler(ctx, args)
	}

	return functiontool.New(cfg, handler)
}

// AdaptToolWithTimeout converts an internal agent.Tool to an ADK tool.Tool
// with an enforced per-call timeout. If timeout <= 0, behaves like AdaptTool.
func AdaptToolWithTimeout(t *agent.Tool, timeout time.Duration) (tool.Tool, error) {
	if timeout <= 0 {
		return AdaptTool(t)
	}

	// Build input schema from parameters
	props := make(map[string]*jsonschema.Schema)
	var required []string

	for name, paramDef := range t.Parameters {
		s := &jsonschema.Schema{}

		if pd, ok := paramDef.(agent.ParameterDef); ok {
			s.Type = pd.Type
			s.Description = pd.Description
			if len(pd.Enum) > 0 {
				s.Enum = make([]any, len(pd.Enum))
				for i, v := range pd.Enum {
					s.Enum[i] = v
				}
			}
			if pd.Required {
				required = append(required, name)
			}
		} else if pdMap, ok := paramDef.(map[string]interface{}); ok {
			if tp, ok := pdMap["type"].(string); ok {
				s.Type = tp
			}
			if d, ok := pdMap["description"].(string); ok {
				s.Description = d
			}
			if r, ok := pdMap["required"].(bool); ok && r {
				required = append(required, name)
			}
		} else {
			s.Type = "string"
		}
		props[name] = s
	}

	inputSchema := &jsonschema.Schema{
		Type:       "object",
		Properties: props,
		Required:   required,
	}

	cfg := functiontool.Config{
		Name:        t.Name,
		Description: t.Description,
		InputSchema: inputSchema,
	}

	handler := func(ctx tool.Context, args map[string]any) (any, error) {
		toolCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		result, err := t.Handler(toolCtx, args)
		if err != nil && toolCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("tool %q timed out after %v", t.Name, timeout)
		}
		return result, err
	}

	return functiontool.New(cfg, handler)
}
