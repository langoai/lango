package smartaccount

import (
	"context"
	"fmt"
	"math/big"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/langoai/lango/internal/smartaccount/policy"
)

type policyShowInfo struct {
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

type policySetResult struct {
	Account      string `json:"account"`
	MaxTxAmount  string `json:"maxTxAmount,omitempty"`
	DailyLimit   string `json:"dailyLimit,omitempty"`
	MonthlyLimit string `json:"monthlyLimit,omitempty"`
}

var loadPolicyShowInfo = func(bootLoader BootLoader) (policyShowInfo, func(), error) {
	boot, err := bootLoader()
	if err != nil {
		return policyShowInfo{}, nil, fmt.Errorf("bootstrap: %w", err)
	}

	deps, err := initSmartAccountDeps(boot)
	if err != nil {
		boot.Close()
		return policyShowInfo{}, nil, err
	}

	ctx := context.Background()
	info, err := deps.manager.Info(ctx)
	if err != nil {
		deps.cleanup()
		boot.Close()
		return policyShowInfo{}, nil, fmt.Errorf("get account info: %w", err)
	}

	result := policyShowInfo{
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

	return result, func() {
		deps.cleanup()
		boot.Close()
	}, nil
}

var updatePolicyLimits = func(bootLoader BootLoader, maxTx, daily, monthly string) (policySetResult, func(), error) {
	parsedMaxTx, parsedDaily, parsedMonthly, err := parsePolicyLimitInputs(maxTx, daily, monthly)
	if err != nil {
		return policySetResult{}, nil, err
	}

	boot, err := bootLoader()
	if err != nil {
		return policySetResult{}, nil, fmt.Errorf("bootstrap: %w", err)
	}

	deps, err := initSmartAccountDeps(boot)
	if err != nil {
		boot.Close()
		return policySetResult{}, nil, err
	}

	ctx := context.Background()
	info, err := deps.manager.Info(ctx)
	if err != nil {
		deps.cleanup()
		boot.Close()
		return policySetResult{}, nil, fmt.Errorf("get account info: %w", err)
	}

	p, _ := deps.policyEngine.GetPolicy(info.Address)
	if p == nil {
		p = &policy.HarnessPolicy{}
	}

	if parsedMaxTx != nil {
		p.MaxTxAmount = parsedMaxTx
	}
	if parsedDaily != nil {
		p.DailyLimit = parsedDaily
	}
	if parsedMonthly != nil {
		p.MonthlyLimit = parsedMonthly
	}

	deps.policyEngine.SetPolicy(info.Address, p)

	result := policySetResult{
		Account: info.Address.Hex(),
	}
	if p.MaxTxAmount != nil {
		result.MaxTxAmount = p.MaxTxAmount.String()
	}
	if p.DailyLimit != nil {
		result.DailyLimit = p.DailyLimit.String()
	}
	if p.MonthlyLimit != nil {
		result.MonthlyLimit = p.MonthlyLimit.String()
	}

	return result, func() {
		deps.cleanup()
		boot.Close()
	}, nil
}

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
	cmd := &cobra.Command{
		Use:           "show",
		Short:         "Show current harness policy configuration",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveTableOrJSONOutput(cmd)
			if err != nil {
				return err
			}
			result, cleanup, err := loadPolicyShowInfo(bootLoader)
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), result)
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
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

	cmd.Flags().String("output", "table", "output format (table|json)")
	return cmd
}

func policySetCmd(bootLoader BootLoader) *cobra.Command {
	var (
		maxTx   string
		daily   string
		monthly string
	)

	cmd := &cobra.Command{
		Use:           "set",
		Short:         "Set harness policy limits",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, cleanup, err := updatePolicyLimits(bootLoader, maxTx, daily, monthly)
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "Policy Updated")
			fmt.Fprintln(w, "--------------")
			fmt.Fprintf(w, "Account:\t%s\n", result.Account)
			if result.MaxTxAmount != "" {
				fmt.Fprintf(w, "Max Tx Amount:\t%s\n", result.MaxTxAmount)
			}
			if result.DailyLimit != "" {
				fmt.Fprintf(w, "Daily Limit:\t%s\n", result.DailyLimit)
			}
			if result.MonthlyLimit != "" {
				fmt.Fprintf(w, "Monthly Limit:\t%s\n", result.MonthlyLimit)
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

func parsePolicyLimit(name, raw string) (*big.Int, error) {
	v, ok := new(big.Int).SetString(raw, 10)
	if !ok {
		return nil, fmt.Errorf("parse %s %q: provide a wei amount (integer)", name, raw)
	}
	return v, nil
}

func parsePolicyLimitInputs(maxTx, daily, monthly string) (*big.Int, *big.Int, *big.Int, error) {
	if maxTx == "" && daily == "" && monthly == "" {
		return nil, nil, nil, fmt.Errorf("provide at least one policy limit (--max-tx, --daily, or --monthly)")
	}

	var parsedMaxTx *big.Int
	var parsedDaily *big.Int
	var parsedMonthly *big.Int
	var err error

	if maxTx != "" {
		parsedMaxTx, err = parsePolicyLimit("max-tx", maxTx)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	if daily != "" {
		parsedDaily, err = parsePolicyLimit("daily", daily)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	if monthly != "" {
		parsedMonthly, err = parsePolicyLimit("monthly", monthly)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	return parsedMaxTx, parsedDaily, parsedMonthly, nil
}
