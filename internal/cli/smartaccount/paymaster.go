package smartaccount

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"text/tabwriter"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	sa "github.com/langoai/lango/internal/smartaccount"
	"github.com/langoai/lango/internal/smartaccount/paymaster"
)

type paymasterStatusInfo struct {
	Enabled          bool   `json:"enabled"`
	Provider         string `json:"provider"`
	Mode             string `json:"mode"`
	RPCURL           string `json:"rpcURL,omitempty"`
	TokenAddress     string `json:"tokenAddress"`
	PaymasterAddress string `json:"paymasterAddress"`
	PolicyID         string `json:"policyId,omitempty"`
	ProviderType     string `json:"providerType,omitempty"`
}

type paymasterApproveResult struct {
	Token     string `json:"token"`
	Paymaster string `json:"paymaster"`
	Amount    string `json:"amount"`
	TxHash    string `json:"txHash"`
}

var loadPaymasterStatus = func(bootLoader BootLoader) (paymasterStatusInfo, func(), error) {
	boot, err := bootLoader()
	if err != nil {
		return paymasterStatusInfo{}, nil, fmt.Errorf("bootstrap: %w", err)
	}

	deps, err := initSmartAccountDeps(boot)
	if err != nil {
		boot.Close()
		return paymasterStatusInfo{}, nil, err
	}

	pmCfg := deps.cfg.Paymaster
	mode := pmCfg.Mode
	if mode == "" {
		mode = "rpc"
	}

	info := paymasterStatusInfo{
		Enabled:          pmCfg.Enabled,
		Provider:         pmCfg.Provider,
		Mode:             mode,
		RPCURL:           pmCfg.RPCURL,
		TokenAddress:     pmCfg.TokenAddress,
		PaymasterAddress: pmCfg.PaymasterAddress,
		PolicyID:         pmCfg.PolicyID,
	}

	if deps.paymasterProv != nil {
		info.ProviderType = deps.paymasterProv.Type()
	}

	return info, func() {
		deps.cleanup()
		boot.Close()
	}, nil
}

var executePaymasterApproval = func(bootLoader BootLoader, amount string) (paymasterApproveResult, func(), error) {
	boot, err := bootLoader()
	if err != nil {
		return paymasterApproveResult{}, nil, fmt.Errorf("bootstrap: %w", err)
	}

	deps, err := initSmartAccountDeps(boot)
	if err != nil {
		boot.Close()
		return paymasterApproveResult{}, nil, err
	}

	pmCfg := deps.cfg.Paymaster
	if !pmCfg.Enabled {
		deps.cleanup()
		boot.Close()
		return paymasterApproveResult{}, nil, fmt.Errorf("paymaster not enabled in config")
	}

	tokenAddr := common.HexToAddress(pmCfg.TokenAddress)
	paymasterAddr := common.HexToAddress(pmCfg.PaymasterAddress)

	var approveAmount *big.Int
	if amount == "max" {
		approveAmount = new(big.Int).Sub(
			new(big.Int).Lsh(big.NewInt(1), 256),
			big.NewInt(1),
		)
	} else {
		var f float64
		if _, scanErr := fmt.Sscanf(amount, "%f", &f); scanErr != nil {
			deps.cleanup()
			boot.Close()
			return paymasterApproveResult{}, nil, fmt.Errorf("parse amount %q: %w", amount, scanErr)
		}
		approveAmount = new(big.Int).SetInt64(int64(f * math.Pow(10, 6)))
	}

	approvalCall := paymaster.NewApprovalCall(tokenAddr, paymasterAddr, approveAmount)

	ctx := context.Background()
	txHash, err := deps.manager.Execute(ctx, []sa.ContractCall{
		{
			Target: approvalCall.TokenAddress,
			Value:  big.NewInt(0),
			Data:   approvalCall.ApproveCalldata,
		},
	})
	if err != nil {
		deps.cleanup()
		boot.Close()
		return paymasterApproveResult{}, nil, fmt.Errorf("execute approval: %w", err)
	}

	result := paymasterApproveResult{
		Token:     tokenAddr.Hex(),
		Paymaster: paymasterAddr.Hex(),
		Amount:    amount,
		TxHash:    txHash,
	}

	return result, func() {
		deps.cleanup()
		boot.Close()
	}, nil
}

func paymasterCmd(bootLoader BootLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "paymaster",
		Short: "Manage ERC-4337 paymaster for gasless USDC transactions",
	}

	cmd.AddCommand(paymasterStatusCmd(bootLoader))
	cmd.AddCommand(paymasterApproveCmd(bootLoader))

	return cmd
}

func paymasterStatusCmd(bootLoader BootLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "status",
		Short:         "Show paymaster configuration and approval status",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveTableOrJSONOutput(cmd)
			if err != nil {
				return err
			}
			info, cleanup, err := loadPaymasterStatus(bootLoader)
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), info)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Paymaster Status")
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "  Enabled:\t%v\n", info.Enabled)
			fmt.Fprintf(w, "  Provider:\t%s\n", info.Provider)
			fmt.Fprintf(w, "  Mode:\t%s\n", info.Mode)
			if info.ProviderType != "" {
				fmt.Fprintf(w, "  Provider Type:\t%s\n", info.ProviderType)
			}
			if info.RPCURL != "" {
				fmt.Fprintf(w, "  RPC URL:\t%s\n", info.RPCURL)
			}
			fmt.Fprintf(w, "  Token:\t%s\n", info.TokenAddress)
			fmt.Fprintf(w, "  Paymaster:\t%s\n", info.PaymasterAddress)
			if info.PolicyID != "" {
				fmt.Fprintf(w, "  Policy ID:\t%s\n", info.PolicyID)
			}
			return w.Flush()
		},
	}

	cmd.Flags().String("output", "table", "output format (table|json)")
	return cmd
}

func paymasterApproveCmd(bootLoader BootLoader) *cobra.Command {
	var (
		output string
		amount string
	)

	cmd := &cobra.Command{
		Use:           "approve",
		Short:         "Approve USDC spending for the paymaster",
		SilenceUsage:  true,
		SilenceErrors: true,
		Long: `Approve the paymaster to spend USDC from your smart account.
This is required before the paymaster can sponsor gas in USDC.

 Examples:
  lango account paymaster approve --amount 1000.00
  lango account paymaster approve --amount max`,
		RunE: func(cmd *cobra.Command, args []string) error {
			output, err := resolveTableOrJSONOutput(cmd)
			if err != nil {
				return err
			}
			result, cleanup, err := executePaymasterApproval(bootLoader, amount)
			if err != nil {
				return err
			}
			if cleanup != nil {
				defer cleanup()
			}

			if output == "json" {
				return printJSON(cmd.OutOrStdout(), result)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Paymaster USDC Approval Submitted")
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintf(w, "  Token:\t%s\n", result.Token)
			fmt.Fprintf(w, "  Paymaster:\t%s\n", result.Paymaster)
			fmt.Fprintf(w, "  Amount:\t%s USDC\n", result.Amount)
			fmt.Fprintf(w, "  Tx Hash:\t%s\n", result.TxHash)
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&output, "output", "table", "output format (table|json)")
	cmd.Flags().StringVar(&amount, "amount", "1000.00", "USDC amount to approve (or 'max' for unlimited)")
	return cmd
}
