package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/logging"
)

// AdaptTools converts all discovered MCP tools from the manager into agent.Tool instances.
// Tool naming follows the convention: mcp__{serverName}__{toolName}
func AdaptTools(mgr *ServerManager, maxOutputTokens int) []*agent.Tool {
	discovered := mgr.AllTools()
	tools := make([]*agent.Tool, 0, len(discovered))

	for _, dt := range discovered {
		conn, ok := mgr.GetConnection(dt.ServerName)
		if !ok {
			continue
		}
		tools = append(tools, AdaptTool(dt, conn, maxOutputTokens))
	}
	return tools
}

// AdaptTool converts a single discovered MCP tool into an agent.Tool.
func AdaptTool(dt DiscoveredTool, conn *ServerConnection, maxOutputTokens int) *agent.Tool {
	tool := dt.Tool
	toolName := fmt.Sprintf("mcp__%s__%s", dt.ServerName, tool.Name)

	// Convert MCP InputSchema to agent.Tool parameters
	params := buildParams(tool.InputSchema)

	// Determine safety level from server config
	safety := parseSafetyLevel(conn.cfg.SafetyLevel)

	t := &agent.Tool{
		Name:        toolName,
		Description: tool.Description,
		Parameters:  params,
		SafetyLevel: safety,
		Handler:     makeHandler(tool.Name, conn, maxOutputTokens),
	}
	t.Capability = agent.ToolCapability{
		Category: "mcp",
		Exposure: agent.ExposureDeferred,
		Activity: agent.ActivityExecute,
	}
	return t
}

func makeHandler(toolName string, conn *ServerConnection, maxOutputTokens int) agent.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		session := conn.Session()
		if session == nil {
			return nil, fmt.Errorf("%w: server %q", ErrNotConnected, conn.Name())
		}

		timeout := conn.timeout()
		callCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		callParams := &sdkmcp.CallToolParams{
			Name:      toolName,
			Arguments: params,
		}

		res, err := session.CallTool(callCtx, callParams)
		if err != nil {
			return nil, fmt.Errorf("%w: %s/%s: %v", ErrToolCallFailed, conn.Name(), toolName, err)
		}

		if res.IsError {
			text := extractText(res.Content)
			return nil, fmt.Errorf("%w: %s/%s: %s", ErrToolCallFailed, conn.Name(), toolName, text)
		}

		result := formatContent(res.Content, maxOutputTokens)
		return map[string]interface{}{
			"result": result,
		}, nil
	}
}

// buildParams converts the MCP tool's InputSchema (any) into agent.Tool parameters.
// The InputSchema from the client is typically a map[string]any with JSON Schema structure.
func buildParams(schema any) map[string]interface{} {
	if schema == nil {
		return nil
	}

	// Convert the schema to a map — it comes as map[string]any from the client SDK.
	var schemaMap map[string]interface{}
	switch v := schema.(type) {
	case map[string]any:
		schemaMap = v
	default:
		// Try JSON round-trip for other types
		data, err := json.Marshal(schema)
		if err != nil {
			return nil
		}
		if err := json.Unmarshal(data, &schemaMap); err != nil {
			return nil
		}
	}

	propsRaw, ok := schemaMap["properties"]
	if !ok {
		return nil
	}
	props, ok := propsRaw.(map[string]interface{})
	if !ok {
		return nil
	}

	// Build required set
	required := make(map[string]bool)
	if reqRaw, ok := schemaMap["required"]; ok {
		if reqSlice, ok := reqRaw.([]interface{}); ok {
			for _, r := range reqSlice {
				if s, ok := r.(string); ok {
					required[s] = true
				}
			}
		}
	}

	params := make(map[string]interface{}, len(props))
	for propName, propRaw := range props {
		propDef, ok := propRaw.(map[string]interface{})
		if !ok {
			params[propName] = map[string]interface{}{
				"type":        "string",
				"description": propName,
			}
			continue
		}

		paramDef := map[string]interface{}{}
		if t, ok := propDef["type"]; ok {
			paramDef["type"] = t
		} else {
			paramDef["type"] = "string"
		}
		if desc, ok := propDef["description"]; ok {
			paramDef["description"] = desc
		}
		if enum, ok := propDef["enum"]; ok {
			paramDef["enum"] = enum
		}
		if required[propName] {
			paramDef["required"] = true
		}
		params[propName] = paramDef
	}

	return params
}

// extractText gets text content from a Content slice for error messages.
func extractText(content []sdkmcp.Content) string {
	var parts []string
	for _, c := range content {
		if tc, ok := c.(*sdkmcp.TextContent); ok {
			parts = append(parts, tc.Text)
		}
	}
	if len(parts) == 0 {
		return "unknown error"
	}
	return strings.Join(parts, "\n")
}

// formatContent processes MCP content into a string result.
func formatContent(content []sdkmcp.Content, maxTokens int) string {
	var parts []string

	for _, c := range content {
		switch v := c.(type) {
		case *sdkmcp.TextContent:
			parts = append(parts, v.Text)
		case *sdkmcp.ImageContent:
			parts = append(parts, fmt.Sprintf("[Image: %s, %d bytes]", v.MIMEType, len(v.Data)))
		case *sdkmcp.AudioContent:
			parts = append(parts, fmt.Sprintf("[Audio: %s]", v.MIMEType))
		default:
			parts = append(parts, fmt.Sprintf("[Content: %T]", c))
		}
	}

	result := strings.Join(parts, "\n")

	// Truncate if exceeding max tokens (approximate: 1 token ~= 4 chars)
	maxChars := maxTokens * 4
	if maxChars > 0 && len(result) > maxChars {
		result = result[:maxChars] + "\n... [truncated]"
		logging.App().Warnw("MCP tool output truncated", "maxTokens", maxTokens, "originalLen", len(result))
	}

	return result
}

// parseSafetyLevel converts a config string to an agent.SafetyLevel.
func parseSafetyLevel(level string) agent.SafetyLevel {
	switch strings.ToLower(level) {
	case "safe":
		return agent.SafetyLevelSafe
	case "moderate":
		return agent.SafetyLevelModerate
	default:
		return agent.SafetyLevelDangerous
	}
}
