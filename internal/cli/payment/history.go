package payment

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/wallet"
)

// Display truncation constants for history table formatting.
const (
	maxHashDisplay    = 14
	truncatedHashLen  = 10
	maxPurposeDisplay = 24
	truncatedPurpLen  = 21
)

func newHistoryCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	var (
		output string
		limit  int
	)

	cmd := &cobra.Command{
		Use:           "history",
		Short:         "Show payment transaction history",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveOutput(cmd)
			if err != nil {
				return err
			}
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}
			defer boot.Close()

			txs, err := paymentHistoryLoader(context.Background(), boot, limit)
			if err != nil {
				return fmt.Errorf("get history: %w", err)
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), map[string]interface{}{
					"transactions": txs,
					"count":        len(txs),
				})
			}

			if len(txs) == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "No transactions found.")
				return err
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "STATUS\tAMOUNT\tTO\tMETHOD\tPURPOSE\tTX HASH\tCREATED")
			for _, tx := range txs {
				hash := tx.TxHash
				if len(hash) > maxHashDisplay {
					hash = hash[:truncatedHashLen] + "..."
				}
				to := tx.To
				if len(to) > maxHashDisplay {
					to = to[:truncatedHashLen] + "..."
				}
				purpose := tx.Purpose
				if len(purpose) > maxPurposeDisplay {
					purpose = purpose[:truncatedPurpLen] + "..."
				}
				method := tx.PaymentMethod
				if method == "" {
					method = "direct"
				}
				fmt.Fprintf(w, "%s\t%s %s\t%s\t%s\t%s\t%s\t%s\n",
					tx.Status,
					tx.Amount,
					wallet.CurrencyUSDC,
					to,
					method,
					purpose,
					hash,
					tx.CreatedAt.Format("2006-01-02 15:04"),
				)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "Output format: table or json")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum number of transactions to show")
	return cmd
}
