package app

import (
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/payment"
	"github.com/langoai/lango/internal/security"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/wallet"
	x402pkg "github.com/langoai/lango/internal/x402"
)

// paymentComponents holds optional blockchain payment components.
type paymentComponents struct {
	wallet    wallet.WalletProvider
	service   *payment.Service
	limiter   wallet.SpendingLimiter
	secrets   *security.SecretsStore
	rpcClient *ethclient.Client
	chainID   int64
}

// initPayment creates the payment components if enabled.
// Follows the same graceful degradation pattern as initGraphStore.
func initPayment(cfg *config.Config, store session.Store, secrets *security.SecretsStore) *paymentComponents {
	if !cfg.Payment.Enabled {
		logger().Info("payment system disabled")
		return nil
	}

	if secrets == nil {
		logger().Warn("payment system requires security.signer, skipping")
		return nil
	}

	entStore, ok := store.(*session.EntStore)
	if !ok {
		logger().Warn("payment system requires EntStore, skipping")
		return nil
	}

	client := entStore.Client()

	// Validate RPC URL before attempting to connect.
	if cfg.Payment.Network.RPCURL == "" {
		logger().Warn("payment RPC URL not configured",
			"fix", "set payment.network.rpcUrl via 'lango config set'")
		return nil
	}

	// Create RPC client for blockchain interaction
	rpcClient, err := ethclient.Dial(cfg.Payment.Network.RPCURL)
	if err != nil {
		logger().Warnw("payment RPC connection failed, skipping", "error", err, "rpcUrl", cfg.Payment.Network.RPCURL)
		return nil
	}

	// Create wallet provider based on configuration
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
		logger().Warnw("unknown wallet provider, using local", "provider", cfg.Payment.WalletProvider)
		wp = wallet.NewLocalWallet(secrets, cfg.Payment.Network.RPCURL, cfg.Payment.Network.ChainID)
	}

	// Create spending limiter
	limiter, err := wallet.NewEntSpendingLimiter(client,
		cfg.Payment.Limits.MaxPerTx,
		cfg.Payment.Limits.MaxDaily,
		cfg.Payment.Limits.AutoApproveBelow,
	)
	if err != nil {
		logger().Warnw("spending limiter init failed, skipping", "error", err)
		return nil
	}

	// Create transaction builder
	builder := payment.NewTxBuilder(rpcClient,
		cfg.Payment.Network.ChainID,
		cfg.Payment.Network.USDCContract,
	)

	// Create payment service
	svc := payment.NewService(wp, limiter, builder, client, rpcClient, cfg.Payment.Network.ChainID)

	logger().Infow("payment system initialized",
		"walletProvider", cfg.Payment.WalletProvider,
		"chainId", cfg.Payment.Network.ChainID,
		"network", wallet.NetworkName(cfg.Payment.Network.ChainID),
		"maxPerTx", cfg.Payment.Limits.MaxPerTx,
		"maxDaily", cfg.Payment.Limits.MaxDaily,
	)

	return &paymentComponents{
		wallet:    wp,
		service:   svc,
		limiter:   limiter,
		secrets:   secrets,
		rpcClient: rpcClient,
		chainID:   cfg.Payment.Network.ChainID,
	}
}

// x402Components holds optional X402 interceptor components.
type x402Components struct {
	interceptor *x402pkg.Interceptor
}

// initX402 creates the X402 interceptor if payment is enabled.
func initX402(cfg *config.Config, secrets *security.SecretsStore, limiter wallet.SpendingLimiter) *x402Components {
	if !cfg.Payment.Enabled {
		return nil
	}
	if secrets == nil {
		logger().Warn("X402 interceptor requires security.signer, skipping")
		return nil
	}

	signerProvider := x402pkg.NewLocalSignerProvider(secrets)

	maxAutoPayAmt := cfg.Payment.Limits.MaxPerTx
	if maxAutoPayAmt == "" {
		maxAutoPayAmt = "1.00"
	}

	x402Cfg := x402pkg.Config{
		Enabled:          true,
		ChainID:          cfg.Payment.Network.ChainID,
		MaxAutoPayAmount: maxAutoPayAmt,
	}

	interceptor := x402pkg.NewInterceptor(signerProvider, limiter, x402Cfg, logger())

	logger().Infow("X402 interceptor configured",
		"chainId", x402Cfg.ChainID,
		"maxAutoPayAmount", maxAutoPayAmt,
	)

	return &x402Components{
		interceptor: interceptor,
	}
}
