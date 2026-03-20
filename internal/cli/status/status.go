// Package status implements the lango status command — a unified status dashboard.
package status

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/cli/tui"
	"github.com/langoai/lango/internal/config"
)

const defaultAddr = "http://localhost:18789"

// NewStatusCmd creates the status command.
func NewStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		outputFmt string
		addr      string
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show unified system status dashboard",
		Long: `Show a unified status dashboard combining health, config, and metrics.

When the server is running, fetches live data from the gateway.
When the server is not running, shows configuration-based status only.

Examples:
  lango status              # Full status dashboard
  lango status --output json  # Machine-readable JSON output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			info := collectStatus(boot.Config, boot.ProfileName, addr)
			info.Version = tui.GetVersion()

			if outputFmt == "json" {
				return printJSON(info)
			}
			fmt.Print(renderDashboard(info))
			return nil
		},
	}

	cmd.Flags().StringVar(&outputFmt, "output", "table", "Output format: table or json")
	cmd.Flags().StringVar(&addr, "addr", defaultAddr, "Gateway address")
	return cmd
}

// StatusInfo holds all collected status data.
type StatusInfo struct {
	Version    string        `json:"version"`
	Profile    string        `json:"profile"`
	ServerUp   bool          `json:"serverUp"`
	Gateway    string        `json:"gateway"`
	Provider   string        `json:"provider"`
	Model      string        `json:"model"`
	Features   []FeatureInfo `json:"features"`
	Channels   []string      `json:"channels"`
	ServerInfo *LiveInfo     `json:"serverInfo,omitempty"`
}

// FeatureInfo describes a feature's status.
type FeatureInfo struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Detail  string `json:"detail,omitempty"`
}

// LiveInfo holds data fetched from a running server.
type LiveInfo struct {
	Healthy bool   `json:"healthy"`
	Uptime  string `json:"uptime,omitempty"`
}

func collectStatus(cfg *config.Config, profile, addr string) StatusInfo {
	info := StatusInfo{
		Profile:  profile,
		Gateway:  fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port),
		Provider: cfg.Agent.Provider,
		Model:    cfg.Agent.Model,
	}

	// Check server health.
	info.ServerUp, info.ServerInfo = probeServer(addr)

	// Collect channels.
	if cfg.Channels.Telegram.Enabled {
		info.Channels = append(info.Channels, "telegram")
	}
	if cfg.Channels.Discord.Enabled {
		info.Channels = append(info.Channels, "discord")
	}
	if cfg.Channels.Slack.Enabled {
		info.Channels = append(info.Channels, "slack")
	}

	// Collect features.
	info.Features = collectFeatures(cfg)

	return info
}

func collectFeatures(cfg *config.Config) []FeatureInfo {
	return []FeatureInfo{
		{"Knowledge", cfg.Knowledge.Enabled, ""},
		{"Embedding & RAG", cfg.Embedding.Provider != "", cfg.Embedding.Provider},
		{"Graph", cfg.Graph.Enabled, ""},
		{"Obs. Memory", cfg.ObservationalMemory.Enabled, ""},
		{"Librarian", cfg.Librarian.Enabled, ""},
		{"Multi-Agent", cfg.Agent.MultiAgent, ""},
		{"Cron", cfg.Cron.Enabled, ""},
		{"Background", cfg.Background.Enabled, ""},
		{"Workflow", cfg.Workflow.Enabled, ""},
		{"MCP", cfg.MCP.Enabled, mcpDetail(cfg)},
		{"P2P", cfg.P2P.Enabled, ""},
		{"Payment", cfg.Payment.Enabled, ""},
		{"Economy", cfg.Economy.Enabled, ""},
		{"A2A", cfg.A2A.Enabled, ""},
		{"RunLedger", cfg.RunLedger.Enabled, ""},
	}
}

func mcpDetail(cfg *config.Config) string {
	if !cfg.MCP.Enabled {
		return ""
	}
	n := len(cfg.MCP.Servers)
	if n == 0 {
		return "no servers"
	}
	return fmt.Sprintf("%d server(s)", n)
}

func probeServer(addr string) (bool, *LiveInfo) {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(addr + "/health")
	if err != nil {
		return false, nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK {
		return true, &LiveInfo{Healthy: true}
	}
	return false, nil
}

func printJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
