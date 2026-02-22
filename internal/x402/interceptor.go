package x402

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"sync"

	x402sdk "github.com/coinbase/x402/go"
	x402http "github.com/coinbase/x402/go/http"
	evmclient "github.com/coinbase/x402/go/mechanisms/evm/exact/client"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/wallet"
)

// Interceptor provides an X402-enabled HTTP client via the Coinbase SDK.
// It lazily initializes the wrapped client and enforces spending limits
// through the SDK's BeforePaymentCreation hook.
type Interceptor struct {
	signerProvider SignerProvider
	limiter        wallet.SpendingLimiter
	config         Config
	logger         *zap.SugaredLogger

	mu         sync.Mutex
	httpClient *http.Client
}

// NewInterceptor creates an X402 interceptor.
func NewInterceptor(sp SignerProvider, limiter wallet.SpendingLimiter, cfg Config, logger *zap.SugaredLogger) *Interceptor {
	return &Interceptor{
		signerProvider: sp,
		limiter:        limiter,
		config:         cfg,
		logger:         logger,
	}
}

// HTTPClient returns an *http.Client that automatically handles HTTP 402 responses
// using the X402 V2 protocol. The client is created lazily and cached.
func (i *Interceptor) HTTPClient(ctx context.Context) (*http.Client, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.httpClient != nil {
		return i.httpClient, nil
	}

	signer, err := i.signerProvider.EvmSigner(ctx)
	if err != nil {
		return nil, fmt.Errorf("create EVM signer: %w", err)
	}

	network := x402sdk.Network(CAIP2Network(i.config.ChainID))
	scheme := evmclient.NewExactEvmScheme(signer)

	// Parse max auto-pay amount for the spending limit hook.
	var maxAutoPayAmt *big.Int
	if i.config.MaxAutoPayAmount != "" {
		maxAutoPayAmt, err = wallet.ParseUSDC(i.config.MaxAutoPayAmount)
		if err != nil {
			return nil, fmt.Errorf("parse maxAutoPayAmount: %w", err)
		}
	}

	// Build the X402 client with a BeforePaymentCreation hook for spending limits.
	opts := []x402sdk.ClientOption{
		x402sdk.WithBeforePaymentCreationHook(func(pctx x402sdk.PaymentCreationContext) (*x402sdk.BeforePaymentCreationHookResult, error) {
			req := pctx.SelectedRequirements
			amountStr := req.GetAmount()

			amount, parseErr := wallet.ParseUSDC(amountStr)
			if parseErr != nil {
				return &x402sdk.BeforePaymentCreationHookResult{
					Abort:  true,
					Reason: fmt.Sprintf("invalid payment amount: %s", amountStr),
				}, nil
			}

			// Check max auto-pay limit.
			if maxAutoPayAmt != nil && amount.Cmp(maxAutoPayAmt) > 0 {
				return &x402sdk.BeforePaymentCreationHookResult{
					Abort:  true,
					Reason: fmt.Sprintf("amount %s exceeds auto-pay limit %s", wallet.FormatUSDC(amount), wallet.FormatUSDC(maxAutoPayAmt)),
				}, nil
			}

			// Check spending limiter.
			if i.limiter != nil {
				hookCtx := pctx.Ctx
				if hookCtx == nil {
					hookCtx = context.Background()
				}
				if limErr := i.limiter.Check(hookCtx, amount); limErr != nil {
					return &x402sdk.BeforePaymentCreationHookResult{
						Abort:  true,
						Reason: limErr.Error(),
					}, nil
				}
			}

			i.logger.Infow("X402 payment approved by spending limiter",
				"amount", amountStr,
				"network", req.GetNetwork(),
			)

			return &x402sdk.BeforePaymentCreationHookResult{Abort: false}, nil
		}),
	}

	x402Client := x402sdk.Newx402Client(opts...)
	x402Client.Register(network, scheme)

	httpX402 := x402http.Newx402HTTPClient(x402Client)
	wrapped := x402http.WrapHTTPClientWithPayment(&http.Client{}, httpX402)

	i.httpClient = wrapped
	i.logger.Infow("X402 interceptor initialized",
		"network", string(network),
		"chainId", i.config.ChainID,
	)

	return i.httpClient, nil
}

// IsEnabled returns whether the interceptor is active.
func (i *Interceptor) IsEnabled() bool {
	return i.config.Enabled
}

// SignerAddress returns the wallet address of the configured signer.
func (i *Interceptor) SignerAddress(ctx context.Context) (string, error) {
	signer, err := i.signerProvider.EvmSigner(ctx)
	if err != nil {
		return "", err
	}
	return signer.Address(), nil
}
