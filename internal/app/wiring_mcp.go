package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/mcp"
)

// mcpComponents holds the results of MCP initialization.
type mcpComponents struct {
	manager *mcp.ServerManager
	tools   []*agent.Tool
}

// initMCP creates the MCP server manager and connects to configured servers.
func initMCP(cfg *config.Config, bus *eventbus.Bus) *mcpComponents {
	if !cfg.MCP.Enabled {
		logger().Info("MCP integration disabled")
		return nil
	}

	// Merge configs from multiple scopes
	merged := mcp.MergedServers(&cfg.MCP)
	if len(merged) == 0 {
		logger().Info("MCP enabled but no servers configured")
		return nil
	}

	// Override profile servers with merged result
	mcpCfg := cfg.MCP
	mcpCfg.Servers = merged

	mgr := mcp.NewServerManager(mcpCfg)

	// Inject OS-level sandbox if enabled.
	if iso := initOSSandbox(cfg); iso != nil {
		if iso.Available() {
			mgr.SetOSIsolator(iso, cfg.DataRoot)
		}
		mgr.SetFailClosed(cfg.Sandbox.FailClosed)
	}
	if bus != nil {
		mgr.SetEventBus(bus)
	}

	// Connect to all servers (best-effort, failures are logged)
	errs := mgr.ConnectAll(context.Background())
	for name, err := range errs {
		logger().Warnw("MCP server failed to connect", "server", name, "error", err)
	}

	// Adapt MCP tools to agent.Tool
	maxTokens := cfg.MCP.MaxOutputTokens
	if maxTokens <= 0 {
		maxTokens = 25000
	}
	tools := mcp.AdaptTools(mgr, maxTokens)

	logger().Infow("MCP integration initialized",
		"servers", mgr.ServerCount(),
		"tools", len(tools),
		"errors", len(errs),
	)

	return &mcpComponents{
		manager: mgr,
		tools:   tools,
	}
}

// buildMCPManagementTools creates meta-tools for managing MCP servers at runtime.
func buildMCPManagementTools(mgr *mcp.ServerManager) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "mcp_status",
			Description: "Show connection status of all MCP servers.",
			Parameters:  nil,
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category: "mcp",
				Activity: agent.ActivityManage,
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				status := mgr.ServerStatus()
				var lines []string
				for name, state := range status {
					lines = append(lines, fmt.Sprintf("%s: %s", name, state))
				}
				if len(lines) == 0 {
					return "No MCP servers configured.", nil
				}
				return strings.Join(lines, "\n"), nil
			},
		},
		{
			Name:        "mcp_tools",
			Description: "List all tools available from MCP servers. Optional: pass 'server' parameter to filter by server name.",
			Parameters: map[string]interface{}{
				"server": map[string]interface{}{
					"type":        "string",
					"description": "Filter tools by server name (optional)",
				},
			},
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category: "mcp",
				Activity: agent.ActivityManage,
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				allTools := mgr.AllTools()
				serverFilter, _ := params["server"].(string)

				var lines []string
				for _, dt := range allTools {
					if serverFilter != "" && dt.ServerName != serverFilter {
						continue
					}
					desc := dt.Tool.Description
					if len(desc) > 80 {
						desc = desc[:80] + "..."
					}
					lines = append(lines, fmt.Sprintf("mcp__%s__%s: %s", dt.ServerName, dt.Tool.Name, desc))
				}
				if len(lines) == 0 {
					return "No MCP tools available.", nil
				}
				return strings.Join(lines, "\n"), nil
			},
		},
	}
}
