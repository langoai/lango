package provenance

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
)

func newAttributionCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attribution",
		Short: "View contribution attribution data",
		Long:  "View raw attribution rows and aggregated contribution reports.",
	}

	cmd.AddCommand(newAttributionShowCmd(bootLoader))
	cmd.AddCommand(newAttributionReportCmd(bootLoader))

	return cmd
}

func newAttributionShowCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "show <session-key>",
		Short: "Show attribution data for a session",
		Args:  cobra.ExactArgs(1),
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

			svcs := loadServices(boot)
			view, err := svcs.attribution.View(cmd.Context(), args[0], 0)
			if err != nil {
				return fmt.Errorf("attribution view: %w", err)
			}

			cmd.Printf("Session:      %s\n", args[0])
			cmd.Printf("Checkpoints:  %d\n", view.Checkpoints)
			cmd.Printf("Total Tokens: in=%d out=%d total=%d\n",
				view.TotalTokens.InputTokens, view.TotalTokens.OutputTokens, view.TotalTokens.TotalTokens)
			for _, row := range view.Attributions {
				cmd.Printf("%s\t%s\t%s\t%s\t+%d\t-%d\n",
					row.CreatedAt.Format("2006-01-02 15:04:05"),
					row.AuthorType,
					row.AuthorID,
					row.Source,
					row.LinesAdded,
					row.LinesRemoved,
				)
			}
			return nil
		},
	}
}

func newAttributionReportCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "report <session-key>",
		Short: "Generate attribution report",
		Args:  cobra.ExactArgs(1),
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

			svcs := loadServices(boot)
			report, err := svcs.attribution.Report(cmd.Context(), args[0])
			if err != nil {
				return fmt.Errorf("attribution report: %w", err)
			}

			cmd.Printf("Session:      %s\n", report.SessionKey)
			cmd.Printf("Checkpoints:  %d\n", report.Checkpoints)
			cmd.Printf("Total Tokens: in=%d out=%d total=%d\n",
				report.TotalTokens.InputTokens, report.TotalTokens.OutputTokens, report.TotalTokens.TotalTokens)
			cmd.Println("\nBy Author:")
			for author, stats := range report.ByAuthor {
				cmd.Printf("  %s\t%s\tfiles=%d\t+%d\t-%d\ttokens=%d\n",
					author, stats.AuthorType, stats.FileCount, stats.LinesAdded, stats.LinesRemoved, stats.TokensUsed.TotalTokens)
			}
			cmd.Println("\nBy File:")
			for path, stats := range report.ByFile {
				cmd.Printf("  %s\tauthors=%d\t+%d\t-%d\n", path, stats.AuthorCount, stats.LinesAdded, stats.LinesRemoved)
			}
			return nil
		},
	}
}
