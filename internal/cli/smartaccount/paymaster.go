package smartaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"text/tabwriter"

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	sa "github.com/langoai/lango/internal/smartaccount"
	"github.com/langoai/lango/internal/smartaccount/paymaster"
)

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
	var output string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show paymaster configuration and approval status",
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

			pmCfg := deps.cfg.Paymaster

			type statusInfo struct {
				Enabled          bool   `json:"enabled"`
				Provider         string `json:"provider"`
				Mode             string `json:"mode"`
				RPCURL           string `json:"rpcURL,omitempty"`
				TokenAddress     string `json:"tokenAddress"`
				PaymasterAddress string `json:"paymasterAddress"`
				PolicyID         string `json:"policyId,omitempty"`
				ProviderType     string `json:"providerType,omitempty"`
			}

			mode := pmCfg.Mode
			if mode == "" {
				mode = "rpc"
			}

			info := statusInfo{
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

			if output == "json" {
				data, marshalErr := json.MarshalIndent(info, "", "  ")
				if marshalErr != nil {
					return fmt.Errorf("marshal json: %w", marshalErr)
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Println("Paymaster Status")
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
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

	cmd.Flags().StringVar(&output, "output", "table", "output format (table|json)")
	return cmd
}

func paymasterApproveCmd(bootLoader BootLoader) *cobra.Command {
	var (
		output string
		amount string
	)

	cmd := &cobra.Command{
		Use:   "approve",
		Short: "Approve USDC spending for the paymaster",
		Long: `Approve the paymaster to spend USDC from your smart account.
This is required before the paymaster can sponsor gas in USDC.

Examples:
  lango account paymaster approve --amount 1000.00
  lango account paymaster approve --amount max`,
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

			pmCfg := deps.cfg.Paymaster
			if !pmCfg.Enabled {
				return fmt.Errorf("paymaster not enabled in config")
			}

			tokenAddr := common.HexToAddress(pmCfg.TokenAddress)
			paymasterAddr := common.HexToAddress(pmCfg.PaymasterAddress)

			// Parse amount (USDC has 6 decimals).
			var approveAmount *big.Int
			if amount == "max" {
				// MaxUint256 for unlimited approval.
				approveAmount = new(big.Int).Sub(
					new(big.Int).Lsh(big.NewInt(1), 256),
					big.NewInt(1),
				)
			} else {
				// Parse as float and convert to 6-decimal integer.
				var f float64
				if _, scanErr := fmt.Sscanf(amount, "%f", &f); scanErr != nil {
					return fmt.Errorf("parse amount %q: %w", amount, scanErr)
				}
				// Convert to smallest unit (6 decimals for USDC).
				approveAmount = new(big.Int).SetInt64(int64(f * math.Pow(10, 6)))
			}

			// Build the approve calldata.
			approvalCall := paymaster.NewApprovalCall(tokenAddr, paymasterAddr, approveAmount)

			// Execute via smart account.
			ctx := context.Background()
			txHash, err := deps.manager.Execute(ctx, []sa.ContractCall{
				{
					Target: approvalCall.TokenAddress,
					Value:  big.NewInt(0),
					Data:   approvalCall.ApproveCalldata,
				},
			})
			if err != nil {
				return fmt.Errorf("execute approval: %w", err)
			}

			type approveResult struct {
				Token     string `json:"token"`
				Paymaster string `json:"paymaster"`
				Amount    string `json:"amount"`
				TxHash    string `json:"txHash"`
			}

			result := approveResult{
				Token:     tokenAddr.Hex(),
				Paymaster: paymasterAddr.Hex(),
				Amount:    amount,
				TxHash:    txHash,
			}

			if output == "json" {
				data, marshalErr := json.MarshalIndent(result, "", "  ")
				if marshalErr != nil {
					return fmt.Errorf("marshal json: %w", marshalErr)
				}
				fmt.Println(string(data))
				return nil
			}

			fmt.Println("Paymaster USDC Approval Submitted")
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
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
