package smartaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/smartaccount/policy"
)

func policyCmd(bootLoader BootLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage session policies",
		Long: `Manage harness policies for smart account session keys.

Examples:
  lango account policy show
  lango account policy set --max-tx "5000000" --daily "50000000" --monthly "500000000"`,
	}

	cmd.AddCommand(policyShowCmd(bootLoader))
	cmd.AddCommand(policySetCmd(bootLoader))

	return cmd
}

func policyShowCmd(bootLoader BootLoader) *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current harness policy configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			deps, err := initSmartAccountDeps(boot)
			if err != nil {
				return err
			}
			defer deps.cleanup()

			// Get account address to look up policy.
			ctx := context.Background()
			info, err := deps.manager.Info(ctx)
			if err != nil {
				return fmt.Errorf("get account info: %w", err)
			}

			type policyInfo struct {
				Account          string   `json:"account"`
				HasPolicy        bool     `json:"hasPolicy"`
				MaxTxAmount      string   `json:"maxTxAmount,omitempty"`
				DailyLimit       string   `json:"dailyLimit,omitempty"`
				MonthlyLimit     string   `json:"monthlyLimit,omitempty"`
				AutoApproveBelow string   `json:"autoApproveBelow,omitempty"`
				AllowedTargets   []string `json:"allowedTargets,omitempty"`
				AllowedFunctions []string `json:"allowedFunctions,omitempty"`
				RiskScore        float64  `json:"requiredRiskScore,omitempty"`
			}

			result := policyInfo{
				Account: info.Address.Hex(),
			}

			p, ok := deps.policyEngine.GetPolicy(info.Address)
			if ok && p != nil {
				result.HasPolicy = true
				if p.MaxTxAmount != nil {
					result.MaxTxAmount = p.MaxTxAmount.String()
				}
				if p.DailyLimit != nil {
					result.DailyLimit = p.DailyLimit.String()
				}
				if p.MonthlyLimit != nil {
					result.MonthlyLimit = p.MonthlyLimit.String()
				}
				if p.AutoApproveBelow != nil {
					result.AutoApproveBelow = p.AutoApproveBelow.String()
				}
				for _, t := range p.AllowedTargets {
					result.AllowedTargets = append(result.AllowedTargets, t.Hex())
				}
				result.AllowedFunctions = p.AllowedFunctions
				result.RiskScore = p.RequiredRiskScore
			}

			if output == "json" {
				data, marshalErr := json.MarshalIndent(result, "", "  ")
				if marshalErr != nil {
					return fmt.Errorf("marshal json: %w", marshalErr)
				}
				fmt.Println(string(data))
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "Harness Policy")
			fmt.Fprintln(w, "==============")
			fmt.Fprintf(w, "Account:\t%s\n", result.Account)
			if !result.HasPolicy {
				fmt.Fprintln(w, "Status:\tNo policy set")
				fmt.Fprintln(w)
				fmt.Fprintln(w, "Use 'lango account policy set' to configure limits.")
			} else {
				fmt.Fprintf(w, "Max Tx Amount:\t%s\n", valueOrNA(result.MaxTxAmount))
				fmt.Fprintf(w, "Daily Limit:\t%s\n", valueOrNA(result.DailyLimit))
				fmt.Fprintf(w, "Monthly Limit:\t%s\n", valueOrNA(result.MonthlyLimit))
				fmt.Fprintf(w, "Auto-Approve Below:\t%s\n", valueOrNA(result.AutoApproveBelow))
				if result.RiskScore > 0 {
					fmt.Fprintf(w, "Required Risk Score:\t%.2f\n", result.RiskScore)
				}
				if len(result.AllowedTargets) > 0 {
					fmt.Fprintf(w, "Allowed Targets:\t%d addresses\n", len(result.AllowedTargets))
				}
				if len(result.AllowedFunctions) > 0 {
					fmt.Fprintf(w, "Allowed Functions:\t%d selectors\n", len(result.AllowedFunctions))
				}
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "output format (table|json)")
	return cmd
}

func policySetCmd(bootLoader BootLoader) *cobra.Command {
	var (
		maxTx   string
		daily   string
		monthly string
	)

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set harness policy limits",
		RunE: func(cmd *cobra.Command, args []string) error {
			boot, err := bootLoader()
			if err != nil {
				return fmt.Errorf("bootstrap: %w", err)
			}
			defer boot.DBClient.Close()

			deps, err := initSmartAccountDeps(boot)
			if err != nil {
				return err
			}
			defer deps.cleanup()

			if maxTx == "" && daily == "" && monthly == "" {
				return fmt.Errorf("provide at least one policy limit (--max-tx, --daily, or --monthly)")
			}

			// Get account address.
			ctx := context.Background()
			info, err := deps.manager.Info(ctx)
			if err != nil {
				return fmt.Errorf("get account info: %w", err)
			}

			// Get existing policy or create new one.
			p, _ := deps.policyEngine.GetPolicy(info.Address)
			if p == nil {
				p = &policy.HarnessPolicy{}
			}

			// Parse and set values.
			if maxTx != "" {
				v, ok := new(big.Int).SetString(maxTx, 10)
				if !ok {
					return fmt.Errorf("parse max-tx %q: provide a wei amount (integer)", maxTx)
				}
				p.MaxTxAmount = v
			}
			if daily != "" {
				v, ok := new(big.Int).SetString(daily, 10)
				if !ok {
					return fmt.Errorf("parse daily %q: provide a wei amount (integer)", daily)
				}
				p.DailyLimit = v
			}
			if monthly != "" {
				v, ok := new(big.Int).SetString(monthly, 10)
				if !ok {
					return fmt.Errorf("parse monthly %q: provide a wei amount (integer)", monthly)
				}
				p.MonthlyLimit = v
			}

			deps.policyEngine.SetPolicy(info.Address, p)

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "Policy Updated")
			fmt.Fprintln(w, "--------------")
			fmt.Fprintf(w, "Account:\t%s\n", info.Address.Hex())
			if p.MaxTxAmount != nil {
				fmt.Fprintf(w, "Max Tx Amount:\t%s\n", p.MaxTxAmount.String())
			}
			if p.DailyLimit != nil {
				fmt.Fprintf(w, "Daily Limit:\t%s\n", p.DailyLimit.String())
			}
			if p.MonthlyLimit != nil {
				fmt.Fprintf(w, "Monthly Limit:\t%s\n", p.MonthlyLimit.String())
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&maxTx, "max-tx", "", "maximum per-transaction amount in wei")
	cmd.Flags().StringVar(&daily, "daily", "", "daily spending limit in wei")
	cmd.Flags().StringVar(&monthly, "monthly", "", "monthly spending limit in wei")

	return cmd
}

// valueOrNA returns the value or "n/a" if empty.
func valueOrNA(s string) string {
	if s == "" {
		return "n/a"
	}
	return s
}

