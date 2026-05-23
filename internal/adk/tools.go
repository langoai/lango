package adk

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/logging"
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

// AdaptStreamingTool converts an internal agent.Tool to an ADK tool.Tool.
// If t.StreamingHandler is non-nil, the returned tool yields partial chunks
// via functiontool.NewStreaming. Otherwise this behaves identically to AdaptTool
// (the streaming intent is opt-in per tool).
//
// In non-live runs, the runtime aggregates yielded chunks into a single result
// string. In Live API runs (Track E), each chunk is delivered to the live session
// as it is produced. Callers that need partial-chunk visibility in non-live mode
// should NOT use this — there is nothing to observe between yields.
func AdaptStreamingTool(t *agent.Tool) (tool.Tool, error) {
	return adaptToolWithOptions(t, "", 0)
}

// buildInputSchema builds a JSON Schema from an agent.Tool's parameter definitions.
// It supports three formats:
//  1. Full JSON Schema from SchemaBuilder.Build(): {"type":"object","properties":{...},"required":[...]}
//  2. Flat map of ParameterDef values: {"name": ParameterDef{...}}
//  3. Flat map of raw maps: {"name": {"type":"string","description":"..."}}
func buildInputSchema(t *agent.Tool) *jsonschema.Schema {
	props := make(map[string]*jsonschema.Schema)
	var required []string

	// Detect full JSON Schema format (from SchemaBuilder.Build()):
	// {"type": "object", "properties": {...}, "required": [...]}
	params := t.Parameters
	if propsRaw, ok := params["properties"]; ok {
		if propsMap, ok := propsRaw.(map[string]interface{}); ok {
			// Extract top-level required array
			if reqRaw, ok := params["required"]; ok {
				if reqSlice, ok := reqRaw.([]string); ok {
					required = reqSlice
				} else if reqIface, ok := reqRaw.([]interface{}); ok {
					for _, r := range reqIface {
						if s, ok := r.(string); ok {
							required = append(required, s)
						}
					}
				}
			}
			// Use nested properties map instead of top-level
			params = propsMap
		}
	}

	for name, paramDef := range params {
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
			s = buildSchemaFromMap(pdMap)
			if r, ok := pdMap["required"].(bool); ok && r {
				required = append(required, name)
			}
		} else {
			s.Type = "string"
		}
		props[name] = s
	}

	return &jsonschema.Schema{
		Type:                 "object",
		Properties:           props,
		Required:             required,
		AdditionalProperties: &jsonschema.Schema{Not: &jsonschema.Schema{}},
	}
}

// buildSchemaFromMap converts a raw map parameter definition to a jsonschema.Schema.
// It handles type, description, enum, items (for arrays), and properties (for nested objects).
func buildSchemaFromMap(m map[string]interface{}) *jsonschema.Schema {
	s := &jsonschema.Schema{}
	if tp, ok := m["type"].(string); ok {
		s.Type = tp
	}
	if d, ok := m["description"].(string); ok {
		s.Description = d
	}
	if enumRaw, ok := m["enum"]; ok {
		if enumStrs, ok := enumRaw.([]string); ok {
			s.Enum = make([]any, len(enumStrs))
			for i, v := range enumStrs {
				s.Enum[i] = v
			}
		} else if enumIface, ok := enumRaw.([]interface{}); ok {
			s.Enum = enumIface
		}
	}
	if s.Type == "array" {
		if itemsRaw, ok := m["items"]; ok {
			if itemsMap, ok := itemsRaw.(map[string]interface{}); ok {
				s.Items = buildSchemaFromMap(itemsMap)
			}
		}
	}
	if s.Type == "object" {
		if propsRaw, ok := m["properties"]; ok {
			if propsMap, ok := propsRaw.(map[string]interface{}); ok {
				props := make(map[string]*jsonschema.Schema, len(propsMap))
				for k, v := range propsMap {
					if vm, ok := v.(map[string]interface{}); ok {
						props[k] = buildSchemaFromMap(vm)
					}
				}
				s.Properties = props
			}
		}
		if reqRaw, ok := m["required"]; ok {
			if reqSlice, ok := reqRaw.([]string); ok {
				s.Required = reqSlice
			} else if reqIface, ok := reqRaw.([]interface{}); ok {
				for _, r := range reqIface {
					if str, ok := r.(string); ok {
						s.Required = append(s.Required, str)
					}
				}
			}
		}
	}
	return s
}

// adaptToolWithOptions is the shared implementation for agent-name-aware tool adaptation.
// When t.StreamingHandler is non-nil, routes through functiontool.NewStreaming;
// otherwise falls back to functiontool.New.
func adaptToolWithOptions(t *agent.Tool, agentName string, timeout time.Duration) (tool.Tool, error) {
	cfg := functiontool.Config{
		Name:        t.Name,
		Description: t.Description,
		InputSchema: buildInputSchema(t),
	}

	toolName := t.Name

	if t.StreamingHandler != nil {
		streamingHandler := func(ctx tool.Context, args map[string]any) iter.Seq2[string, error] {
			callCtx := injectCallContext(ctx, agentName)
			if timeout > 0 {
				var cancel context.CancelFunc
				callCtx, cancel = context.WithTimeout(callCtx, timeout)
				// NOTE: cancel is called when the consumer stops ranging over the iterator
				// or when the timeout fires. Defer is not used because the iter.Seq2 may
				// outlive this function's stack frame.
				return func(yield func(string, error) bool) {
					defer cancel()
					for chunk, err := range t.StreamingHandler(callCtx, args) {
						if err != nil && callCtx.Err() == context.DeadlineExceeded {
							err = fmt.Errorf("tool %q timed out after %v", toolName, timeout)
						}
						if !yield(chunk, err) {
							return
						}
					}
				}
			}
			return t.StreamingHandler(callCtx, args)
		}
		return functiontool.NewStreaming[map[string]any](cfg, streamingHandler)
	}

	handler := func(ctx tool.Context, args map[string]any) (any, error) {
		callCtx := injectCallContext(ctx, agentName)

		var result any
		var err error

		if timeout > 0 {
			var cancel context.CancelFunc
			callCtx, cancel = context.WithTimeout(callCtx, timeout)
			defer cancel()

			result, err = t.Handler(callCtx, args)
			if err != nil && callCtx.Err() == context.DeadlineExceeded {
				err = fmt.Errorf("tool %q timed out after %v", toolName, timeout)
			}
		} else {
			result, err = t.Handler(callCtx, args)
		}

		if err != nil {
			logging.Agent().Warnw("tool call failed",
				"tool", toolName,
				"agent", agentName,
				"error", err,
			)
		}

		return result, err
	}

	return functiontool.New(cfg, handler)
}

// injectCallContext wires the agent name into the call context so downstream
// hooks/middleware can identify the owning agent.
func injectCallContext(ctx tool.Context, agentName string) context.Context {
	var callCtx context.Context = ctx
	if agentName != "" {
		callCtx = ctxkeys.WithAgentName(ctx, agentName)
	}
	return callCtx
}
