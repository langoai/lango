// Package payment provides CLI commands for blockchain payment management.
package payment

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"

	"github.com/langowarny/lango/internal/bootstrap"
	"github.com/langowarny/lango/internal/config"
	"github.com/langowarny/lango/internal/payment"
	"github.com/langowarny/lango/internal/security"
	"github.com/langowarny/lango/internal/session"
	"github.com/langowarny/lango/internal/wallet"
)

// paymentDeps holds lazily-initialized payment dependencies.
type paymentDeps struct {
	service *payment.Service
	limiter *wallet.EntSpendingLimiter
	config  *config.PaymentConfig
	cleanup func()
}

// NewPaymentCmd creates the payment command with lazy bootstrap loading.
func NewPaymentCmd(bootLoader func() (*bootstrap.Result, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payment",
		Short: "Manage blockchain payments",
		Long:  "View balances, transaction history, spending limits, and send USDC payments on Base L2.",
	}

	cmd.AddCommand(newBalanceCmd(bootLoader))
	cmd.AddCommand(newHistoryCmd(bootLoader))
	cmd.AddCommand(newLimitsCmd(bootLoader))
	cmd.AddCommand(newInfoCmd(bootLoader))
	cmd.AddCommand(newSendCmd(bootLoader))

	return cmd
}

// initPaymentDeps creates payment components from a bootstrap result.
// Unlike wiring.go which degrades gracefully, CLI returns errors so the user
// knows exactly what is misconfigured.
func initPaymentDeps(boot *bootstrap.Result) (*paymentDeps, error) {
	cfg := boot.Config
	if !cfg.Payment.Enabled {
		return nil, fmt.Errorf("payment system is not enabled (set payment.enabled = true)")
	}

	// Build secrets store for wallet key management.
	ctx := context.Background()
	registry := security.NewKeyRegistry(boot.DBClient)
	if _, err := registry.RegisterKey(ctx, "default", "local", security.KeyTypeEncryption); err != nil {
		return nil, fmt.Errorf("register default key: %w", err)
	}
	secrets := security.NewSecretsStore(boot.DBClient, registry, boot.Crypto)

	// Get ent client for payment records.
	client := session.NewEntStoreWithClient(boot.DBClient).Client()

	// Create RPC client for blockchain interaction.
	rpcClient, err := ethclient.Dial(cfg.Payment.Network.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("connect to RPC %q: %w", cfg.Payment.Network.RPCURL, err)
	}

	// Create wallet provider based on configuration.
	var wp wallet.WalletProvider
	switch cfg.Payment.WalletProvider {
	case "local":
		wp = wallet.NewLocalWallet(secrets, cfg.Payment.Network.RPCURL, cfg.Payment.Network.ChainID)
	case "rpc":
		wp = wallet.NewRPCWallet()
	case "composite":
		local := wallet.NewLocalWallet(secrets, cfg.Payment.Network.RPCURL, cfg.Payment.Network.ChainID)
		rpc := wallet.NewRPCWallet()
		wp = wallet.NewCompositeWallet(rpc, local, nil)
	default:
		wp = wallet.NewLocalWallet(secrets, cfg.Payment.Network.RPCURL, cfg.Payment.Network.ChainID)
	}

	// Create spending limiter.
	limiter, err := wallet.NewEntSpendingLimiter(client,
		cfg.Payment.Limits.MaxPerTx,
		cfg.Payment.Limits.MaxDaily,
	)
	if err != nil {
		rpcClient.Close()
		return nil, fmt.Errorf("init spending limiter: %w", err)
	}

	// Create transaction builder.
	builder := payment.NewTxBuilder(rpcClient,
		cfg.Payment.Network.ChainID,
		cfg.Payment.Network.USDCContract,
	)

	// Create payment service.
	svc := payment.NewService(wp, limiter, builder, client, rpcClient, cfg.Payment.Network.ChainID)

	return &paymentDeps{
		service: svc,
		limiter: limiter,
		config:  &cfg.Payment,
		cleanup: func() {
			rpcClient.Close()
		},
	}, nil
}
