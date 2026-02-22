package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/langoai/lango/internal/config"
	"github.com/spf13/cobra"
)

func newStatusCmd(cfgLoader func() (*config.Config, error)) *cobra.Command {
	var (
		sessionKey string
		jsonOutput bool
	)

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show observational memory status for a session",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cfgLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			store, memStore, cleanup, err := initMemoryStore(cfg)
			if err != nil {
				return err
			}
			_ = store
			defer cleanup()

			ctx := context.Background()

			obs, err := memStore.ListObservations(ctx, sessionKey)
			if err != nil {
				return fmt.Errorf("list observations: %w", err)
			}

			refs, err := memStore.ListReflections(ctx, sessionKey)
			if err != nil {
				return fmt.Errorf("list reflections: %w", err)
			}

			obsTokens := 0
			for _, o := range obs {
				obsTokens += o.TokenCount
			}
			refTokens := 0
			for _, r := range refs {
				refTokens += r.TokenCount
			}

			type statusOutput struct {
				Observations              int    `json:"observations"`
				Reflections               int    `json:"reflections"`
				ObservationTokens         int    `json:"observation_tokens"`
				ReflectionTokens          int    `json:"reflection_tokens"`
				Enabled                   bool   `json:"enabled"`
				Provider                  string `json:"provider,omitempty"`
				Model                     string `json:"model,omitempty"`
				MessageTokenThreshold     int    `json:"message_token_threshold"`
				ObservationTokenThreshold int    `json:"observation_token_threshold"`
				MaxMessageTokenBudget     int    `json:"max_message_token_budget"`
			}

			s := statusOutput{
				Observations:              len(obs),
				Reflections:               len(refs),
				ObservationTokens:         obsTokens,
				ReflectionTokens:          refTokens,
				Enabled:                   cfg.ObservationalMemory.Enabled,
				Provider:                  cfg.ObservationalMemory.Provider,
				Model:                     cfg.ObservationalMemory.Model,
				MessageTokenThreshold:     cfg.ObservationalMemory.MessageTokenThreshold,
				ObservationTokenThreshold: cfg.ObservationalMemory.ObservationTokenThreshold,
				MaxMessageTokenBudget:     cfg.ObservationalMemory.MaxMessageTokenBudget,
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(s)
			}

			fmt.Printf("Observational Memory Status (session: %s)\n", sessionKey)
			fmt.Printf("  Enabled:                      %v\n", s.Enabled)
			if s.Provider != "" {
				fmt.Printf("  Provider:                     %s\n", s.Provider)
			}
			if s.Model != "" {
				fmt.Printf("  Model:                        %s\n", s.Model)
			}
			fmt.Printf("  Observations:                 %d (%d tokens)\n",
				s.Observations, s.ObservationTokens)
			fmt.Printf("  Reflections:                  %d (%d tokens)\n",
				s.Reflections, s.ReflectionTokens)
			fmt.Printf("  Message Token Threshold:      %d\n", s.MessageTokenThreshold)
			fmt.Printf("  Observation Token Threshold:  %d\n", s.ObservationTokenThreshold)
			fmt.Printf("  Max Message Token Budget:     %d\n", s.MaxMessageTokenBudget)

			return nil
		},
	}

	cmd.Flags().StringVar(&sessionKey, "session", "", "Session key (required)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	_ = cmd.MarkFlagRequired("session")

	return cmd
}
