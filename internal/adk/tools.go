package adk

import (
	"context"
	"fmt"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ctxkeys"
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

// AdaptToolForAgent converts an internal agent.Tool to an ADK tool.Tool and
// injects the given agentName into the context for every handler invocation.
// This allows downstream hooks and middleware to identify which agent owns the tool call.
func AdaptToolForAgent(t *agent.Tool, agentName string) (tool.Tool, error) {
	return adaptToolWithOptions(t, agentName, 0)
}

// AdaptToolForAgentWithTimeout combines agent name injection with a per-call timeout.
func AdaptToolForAgentWithTimeout(t *agent.Tool, agentName string, timeout time.Duration) (tool.Tool, error) {
	return adaptToolWithOptions(t, agentName, timeout)
}

// adaptToolWithOptions is the shared implementation for agent-name-aware tool adaptation.
func adaptToolWithOptions(t *agent.Tool, agentName string, timeout time.Duration) (tool.Tool, error) {
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
		// Inject agent name into context so hooks/middleware can identify the owning agent.
		var callCtx context.Context = ctx
		if agentName != "" {
			callCtx = ctxkeys.WithAgentName(ctx, agentName)
		}

		if timeout > 0 {
			var cancel context.CancelFunc
			callCtx, cancel = context.WithTimeout(callCtx, timeout)
			defer cancel()

			result, err := t.Handler(callCtx, args)
			if err != nil && callCtx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("tool %q timed out after %v", t.Name, timeout)
			}
			return result, err
		}

		return t.Handler(callCtx, args)
	}

	return functiontool.New(cfg, handler)
}
