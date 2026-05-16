package agent

import (
	"fmt"

	"github.com/langoai/lango/internal/agentregistry"
	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

var loadAgentRegistryCounts = func(cfg *config.Config) (builtinCount, userCount, activeCount int) {
	reg := agentregistry.New()
	embeddedStore := agentregistry.NewEmbeddedStore()
	_ = reg.LoadFromStore(embeddedStore)
	builtinCount = len(reg.All())

	if cfg.Agent.AgentsDir != "" {
		userStore := agentregistry.NewFileStore(cfg.Agent.AgentsDir)
		_ = reg.LoadFromStore(userStore)
		userCount = len(reg.All()) - builtinCount
	}

	return builtinCount, userCount, len(reg.Active())
}

func newStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:           "status",
		Short:         "Show agent mode, configuration, and registry info",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			mode := "single"
			if cfg.Agent.MultiAgent {
				mode = "multi-agent"
			}

			type registryInfo struct {
				Builtin  int    `json:"builtin"`
				User     int    `json:"user"`
				Active   int    `json:"active"`
				AgentDir string `json:"agents_dir,omitempty"`
			}

			type statusOutput struct {
				Mode                   string       `json:"mode"`
				TeammateRuntime        string       `json:"teammate_runtime,omitempty"`
				Provider               string       `json:"provider"`
				Model                  string       `json:"model"`
				MultiAgent             bool         `json:"multi_agent"`
				A2AEnabled             bool         `json:"a2a_enabled"`
				A2ABaseURL             string       `json:"a2a_base_url,omitempty"`
				A2AAgent               string       `json:"a2a_agent_name,omitempty"`
				RemoteAgents           int          `json:"remote_agents"`
				MaxTurns               int          `json:"max_turns"`
				ErrorCorrectionEnabled bool         `json:"error_correction_enabled"`
				MaxDelegationRounds    int          `json:"max_delegation_rounds,omitempty"`
				P2PEnabled             bool         `json:"p2p_enabled"`
				HooksEnabled           bool         `json:"hooks_enabled"`
				Registry               registryInfo `json:"registry"`
			}

			// Compute effective defaults.
			maxTurns := cfg.Agent.MaxTurns
			if maxTurns <= 0 {
				maxTurns = 50
				if cfg.Agent.MultiAgent {
					maxTurns = 75
				}
			}
			errorCorrection := true
			if cfg.Agent.ErrorCorrectionEnabled != nil {
				errorCorrection = *cfg.Agent.ErrorCorrectionEnabled
			}
			maxDelegation := cfg.Agent.MaxDelegationRounds
			if maxDelegation <= 0 {
				maxDelegation = 10
			}
			teammateRuntime := ""
			if cfg.Agent.MultiAgent && cfg.Background.Enabled {
				teammateRuntime = "dynamic-v1"
			}
			teammateRuntimeHint := ""
			if cfg.Agent.MultiAgent && !cfg.Background.Enabled {
				teammateRuntimeHint = "Enable background.enabled to report dynamic-v1 teammate runtime."
			}

			builtinCount, userCount, activeCount := loadAgentRegistryCounts(cfg)

			s := statusOutput{
				Mode:                   mode,
				TeammateRuntime:        teammateRuntime,
				Provider:               cfg.Agent.Provider,
				Model:                  cfg.Agent.Model,
				MultiAgent:             cfg.Agent.MultiAgent,
				A2AEnabled:             cfg.A2A.Enabled,
				RemoteAgents:           len(cfg.A2A.RemoteAgents),
				MaxTurns:               maxTurns,
				ErrorCorrectionEnabled: errorCorrection,
				MaxDelegationRounds:    maxDelegation,
				P2PEnabled:             cfg.P2P.Enabled,
				HooksEnabled:           cfg.Hooks.Enabled,
				Registry: registryInfo{
					Builtin:  builtinCount,
					User:     userCount,
					Active:   activeCount,
					AgentDir: cfg.Agent.AgentsDir,
				},
			}
			if cfg.A2A.Enabled {
				s.A2ABaseURL = cfg.A2A.BaseURL
				s.A2AAgent = cfg.A2A.AgentName
			}

			if output == "json" {
				return printPrettyJSON(cmd.OutOrStdout(), s)
			}

			writer := cmd.OutOrStdout()
			fmt.Fprintf(writer, "Agent Status\n")
			fmt.Fprintf(writer, "  Mode:              %s\n", s.Mode)
			fmt.Fprintf(writer, "  Provider:          %s\n", s.Provider)
			fmt.Fprintf(writer, "  Model:             %s\n", s.Model)
			fmt.Fprintf(writer, "  Multi-Agent:       %v\n", s.MultiAgent)
			if s.TeammateRuntime != "" {
				fmt.Fprintf(writer, "  Teammate Runtime:  %s\n", s.TeammateRuntime)
			} else if teammateRuntimeHint != "" {
				fmt.Fprintf(writer, "  Teammate Runtime:  unavailable\n")
				fmt.Fprintf(writer, "  Runtime Hint:      %s\n", teammateRuntimeHint)
			}
			fmt.Fprintf(writer, "  Max Turns:         %d\n", s.MaxTurns)
			fmt.Fprintf(writer, "  Error Correction:  %v\n", s.ErrorCorrectionEnabled)
			if s.MultiAgent {
				fmt.Fprintf(writer, "  Delegation Rounds: %d\n", s.MaxDelegationRounds)
			}
			fmt.Fprintf(writer, "  A2A Enabled:       %v\n", s.A2AEnabled)
			if s.A2AEnabled {
				fmt.Fprintf(writer, "  A2A Base URL:      %s\n", s.A2ABaseURL)
				fmt.Fprintf(writer, "  A2A Agent:         %s\n", s.A2AAgent)
				fmt.Fprintf(writer, "  Remote Agents:     %d\n", s.RemoteAgents)
			}
			fmt.Fprintf(writer, "  P2P Enabled:       %v\n", s.P2PEnabled)
			fmt.Fprintf(writer, "  Hooks Enabled:     %v\n", s.HooksEnabled)
			fmt.Fprintln(writer)
			fmt.Fprintf(writer, "Agent Registry\n")
			fmt.Fprintf(writer, "  Builtin Agents:    %d\n", s.Registry.Builtin)
			fmt.Fprintf(writer, "  User Agents:       %d\n", s.Registry.User)
			fmt.Fprintf(writer, "  Active Agents:     %d\n", s.Registry.Active)
			if s.Registry.AgentDir != "" {
				fmt.Fprintf(writer, "  Agents Dir:        %s\n", s.Registry.AgentDir)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")

	return cmd
}
