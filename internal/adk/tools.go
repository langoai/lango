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

// AdaptTool converts an internal agent.Tool to an ADK tool.Tool.
func AdaptTool(t *agent.Tool) (tool.Tool, error) {
	return adaptToolWithOptions(t, "", 0)
}

// AdaptToolWithTimeout converts an internal agent.Tool to an ADK tool.Tool
// with an enforced per-call timeout. If timeout <= 0, behaves like AdaptTool.
func AdaptToolWithTimeout(t *agent.Tool, timeout time.Duration) (tool.Tool, error) {
	return adaptToolWithOptions(t, "", timeout)
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

// buildInputSchema builds a JSON Schema from an agent.Tool's parameter definitions.
func buildInputSchema(t *agent.Tool) *jsonschema.Schema {
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

	return &jsonschema.Schema{
		Type:       "object",
		Properties: props,
		Required:   required,
	}
}

// adaptToolWithOptions is the shared implementation for agent-name-aware tool adaptation.
func adaptToolWithOptions(t *agent.Tool, agentName string, timeout time.Duration) (tool.Tool, error) {
	cfg := functiontool.Config{
		Name:        t.Name,
		Description: t.Description,
		InputSchema: buildInputSchema(t),
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
