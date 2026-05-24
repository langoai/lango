package provenance

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	provenancepkg "github.com/langoai/lango/internal/provenance"
	"github.com/langoai/lango/internal/runledger"
)

func provenanceRunLedgerStore(boot *bootstrap.Result) runledger.RunLedgerStore {
	if boot != nil && boot.Storage != nil {
		if store := boot.Storage.RunLedger(); store != nil {
			return store
		}
	}
	return nil
}

func newCheckpointCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkpoint",
		Short: "Manage provenance checkpoints",
	}

	cmd.AddCommand(newCheckpointListCmd(bootLoader))
	cmd.AddCommand(newCheckpointCreateCmd(bootLoader))
	cmd.AddCommand(newCheckpointShowCmd(bootLoader))

	return cmd
}

func newCheckpointListCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		runID      string
		sessionKey string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List checkpoints",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return err
			}
			defer boot.Close()

			if isProvenanceDisabled(boot, cmd) {
				return nil
			}

			store := loadServices(boot).checkpoints
			ctx := context.Background()

			var checkpoints []provenancepkg.Checkpoint
			if runID != "" {
				checkpoints, err = store.ListByRun(ctx, runID)
			} else if sessionKey != "" {
				checkpoints, err = store.ListBySession(ctx, sessionKey, 50)
			} else {
				cmd.Println("Specify --run or --session to filter checkpoints.")
				return nil
			}
			if err != nil {
				return fmt.Errorf("list checkpoints: %w", err)
			}

			if len(checkpoints) == 0 {
				cmd.Println("No checkpoints found.")
				return nil
			}

			for _, cp := range checkpoints {
				cmd.Printf("%s\t%s\t%s\tseq=%d\t%s\n",
					cp.ID[:8], cp.Trigger, cp.Label, cp.JournalSeq, cp.CreatedAt.Format(dateTimeFormat))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&runID, "run", "", "Filter by run ID")
	cmd.Flags().StringVar(&sessionKey, "session", "", "Filter by session key")

	return cmd
}

func newCheckpointCreateCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var runID string

	cmd := &cobra.Command{
		Use:   "create <label>",
		Short: "Create a manual checkpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if runID == "" {
				return fmt.Errorf("--run is required")
			}

			boot, err := bootLoader()
			if err != nil {
				return err
			}
			defer boot.Close()

			if isProvenanceDisabled(boot, cmd) {
				return nil
			}

			svcs := loadServices(boot)
			cpStore := svcs.checkpoints
			ledgerStore := provenanceRunLedgerStore(boot)
			if ledgerStore == nil {
				return fmt.Errorf("runledger store unavailable")
			}
			svc := provenancepkg.NewCheckpointService(cpStore, ledgerStore, boot.Config.Provenance.Checkpoints)

			cp, err := svc.CreateManual(context.Background(), "", runID, args[0])
			if err != nil {
				return fmt.Errorf("create checkpoint: %w", err)
			}

			cmd.Printf("Checkpoint created: %s (seq=%d)\n", cp.ID, cp.JournalSeq)
			return nil
		},
	}

	cmd.Flags().StringVar(&runID, "run", "", "Run ID (required)")
	_ = cmd.MarkFlagRequired("run")

	return cmd
}

func newCheckpointShowCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	return &cobra.Command{
		Use:   "show <checkpoint-id>",
		Short: "Show checkpoint details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return err
			}
			defer boot.Close()

			if isProvenanceDisabled(boot, cmd) {
				return nil
			}

			store := loadServices(boot).checkpoints
			cp, err := store.GetCheckpoint(context.Background(), args[0])
			if err != nil {
				return fmt.Errorf("get checkpoint: %w", err)
			}

			cmd.Printf("ID:          %s\n", cp.ID)
			cmd.Printf("Label:       %s\n", cp.Label)
			cmd.Printf("Trigger:     %s\n", cp.Trigger)
			cmd.Printf("Session:     %s\n", cp.SessionKey)
			cmd.Printf("Run:         %s\n", cp.RunID)
			cmd.Printf("Journal Seq: %d\n", cp.JournalSeq)
			cmd.Printf("Git Ref:     %s\n", cp.GitRef)
			cmd.Printf("Created:     %s\n", cp.CreatedAt.Format(dateTimeFormat))
			return nil
		},
	}
}
