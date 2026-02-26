package agent

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

func newStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show agent mode and configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			mode := "single"
			if cfg.Agent.MultiAgent {
				mode = "multi-agent"
			}

			type statusOutput struct {
				Mode                   string `json:"mode"`
				Provider               string `json:"provider"`
				Model                  string `json:"model"`
				MultiAgent             bool   `json:"multi_agent"`
				A2AEnabled             bool   `json:"a2a_enabled"`
				A2ABaseURL             string `json:"a2a_base_url,omitempty"`
				A2AAgent               string `json:"a2a_agent_name,omitempty"`
				MaxTurns               int    `json:"max_turns"`
				ErrorCorrectionEnabled bool   `json:"error_correction_enabled"`
				MaxDelegationRounds    int    `json:"max_delegation_rounds,omitempty"`
			}

			// Compute effective defaults.
			maxTurns := cfg.Agent.MaxTurns
			if maxTurns <= 0 {
				maxTurns = 25
			}
			errorCorrection := true
			if cfg.Agent.ErrorCorrectionEnabled != nil {
				errorCorrection = *cfg.Agent.ErrorCorrectionEnabled
			}
			maxDelegation := cfg.Agent.MaxDelegationRounds
			if maxDelegation <= 0 {
				maxDelegation = 10
			}

			s := statusOutput{
				Mode:                   mode,
				Provider:               cfg.Agent.Provider,
				Model:                  cfg.Agent.Model,
				MultiAgent:             cfg.Agent.MultiAgent,
				A2AEnabled:             cfg.A2A.Enabled,
				MaxTurns:               maxTurns,
				ErrorCorrectionEnabled: errorCorrection,
				MaxDelegationRounds:    maxDelegation,
			}
			if cfg.A2A.Enabled {
				s.A2ABaseURL = cfg.A2A.BaseURL
				s.A2AAgent = cfg.A2A.AgentName
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(s)
			}

			fmt.Printf("Agent Status\n")
			fmt.Printf("  Mode:              %s\n", s.Mode)
			fmt.Printf("  Provider:          %s\n", s.Provider)
			fmt.Printf("  Model:             %s\n", s.Model)
			fmt.Printf("  Multi-Agent:       %v\n", s.MultiAgent)
			fmt.Printf("  Max Turns:         %d\n", s.MaxTurns)
			fmt.Printf("  Error Correction:  %v\n", s.ErrorCorrectionEnabled)
			if s.MultiAgent {
				fmt.Printf("  Delegation Rounds: %d\n", s.MaxDelegationRounds)
			}
			fmt.Printf("  A2A Enabled:       %v\n", s.A2AEnabled)
			if cfg.A2A.Enabled {
				fmt.Printf("  A2A Base URL:      %s\n", s.A2ABaseURL)
				fmt.Printf("  A2A Agent:         %s\n", s.A2AAgent)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
