package memory

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/config"
)

func newAgentsCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Show agent memory configuration and status",
		Long:  "Display agent memory system configuration. Agent memories are persistent, retained across restarts.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			type agentMemoryStatus struct {
				Enabled bool   `json:"enabled"`
				Note    string `json:"note"`
			}

			out := agentMemoryStatus{
				Enabled: cfg.AgentMemory.Enabled,
				Note:    "Agent memories are persistent. Use the server API to list active agent memories.",
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			fmt.Println("Agent Memory")
			fmt.Printf("  Enabled: %v\n", out.Enabled)
			fmt.Println()
			fmt.Println("  Note: Agent memories are persistent, retained across restarts.")
			fmt.Println("  Use the server API to inspect active agent memories.")

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newAgentCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "agent <name>",
		Short: "Show memories for a specific agent",
		Long:  "Display stored memories for a named agent. Agent memories are persistent, retained across restarts.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			if !cfg.AgentMemory.Enabled {
				return fmt.Errorf("agent memory is not enabled (set agentMemory.enabled = true)")
			}

			agentName := args[0]

			type agentInfo struct {
				AgentName string `json:"agent_name"`
				Note      string `json:"note"`
			}

			out := agentInfo{
				AgentName: agentName,
				Note:      "Agent memories are persistent. Connect to the running server to query memories.",
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			fmt.Printf("Agent Memory: %s\n", agentName)
			fmt.Println()
			fmt.Println("  Agent memories are persistent, retained across restarts.")
			fmt.Println("  Connect to the running server API to query memories for this agent.")

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
