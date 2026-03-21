// Package provenance provides CLI commands for session provenance management.
package provenance

import (
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

// NewProvenanceCmd creates the provenance command group with lazy bootstrap loading.
func NewProvenanceCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provenance",
		Short: "Session provenance: checkpoints, session tree, attribution",
		Long:  "Manage session provenance data including checkpoints (persistent), session trees (not yet implemented), and contribution attribution (not yet implemented).",
	}

	cmd.AddCommand(newStatusCmd(bootLoader))
	cmd.AddCommand(newCheckpointCmd(bootLoader))
	cmd.AddCommand(newSessionCmd(bootLoader))
	cmd.AddCommand(newAttributionCmd(bootLoader))

	return cmd
}

func newStatusCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show provenance configuration and state",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return err
			}
			defer boot.DBClient.Close()

			cfg := boot.Config.Provenance
			cmd.Printf("Provenance Configuration:\n")
			cmd.Printf("  Enabled:              %v\n", cfg.Enabled)
			cmd.Printf("  Auto on Step Complete: %v\n", cfg.Checkpoints.AutoOnStepComplete)
			cmd.Printf("  Auto on Policy:       %v\n", cfg.Checkpoints.AutoOnPolicy)
			cmd.Printf("  Max per Session:      %d\n", cfg.Checkpoints.MaxPerSession)
			cmd.Printf("  Retention Days:       %d\n", cfg.Checkpoints.RetentionDays)
			return nil
		},
	}
}
