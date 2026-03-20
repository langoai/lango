package provenance

import (
	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

func newAttributionCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attribution",
		Short: "View contribution attribution data",
		Long:  "View and generate attribution reports (Phase 3 — not yet implemented).",
	}

	cmd.AddCommand(newAttributionShowCmd(bootLoader))
	cmd.AddCommand(newAttributionReportCmd(bootLoader))

	return cmd
}

func newAttributionShowCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show attribution data for a session",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return err
			}
			defer boot.DBClient.Close()

			if !boot.Config.Provenance.Enabled {
				cmd.Println("Provenance is disabled. Enable with: lango config set provenance.enabled true")
				return nil
			}

			cmd.Println("Attribution show: not yet implemented (Phase 3)")
			return nil
		},
	}
}

func newAttributionReportCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "report",
		Short: "Generate attribution report",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return err
			}
			defer boot.DBClient.Close()

			if !boot.Config.Provenance.Enabled {
				cmd.Println("Provenance is disabled. Enable with: lango config set provenance.enabled true")
				return nil
			}

			cmd.Println("Attribution report: not yet implemented (Phase 3)")
			return nil
		},
	}
}
