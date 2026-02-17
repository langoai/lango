package x402

import (
	"context"
	"fmt"
	"math/big"

	"go.uber.org/zap"

	"github.com/langowarny/lango/internal/payment"
	"github.com/langowarny/lango/internal/wallet"
)

// Interceptor handles automatic payment for HTTP 402 responses.
type Interceptor struct {
	service      *payment.Service
	maxAutoPayAmt *big.Int
	enabled       bool
	logger        *zap.SugaredLogger
}

// NewInterceptor creates an X402 interceptor.
func NewInterceptor(svc *payment.Service, maxAutoPayAmount string, enabled bool, logger *zap.SugaredLogger) (*Interceptor, error) {
	maxAmt, err := wallet.ParseUSDC(maxAutoPayAmount)
	if err != nil {
		return nil, fmt.Errorf("parse maxAutoPayAmount: %w", err)
	}

	return &Interceptor{
		service:       svc,
		maxAutoPayAmt: maxAmt,
		enabled:       enabled,
		logger:        logger,
	}, nil
}

// HandleChallenge processes an X402 challenge by making a payment and returning
// the payment proof header value.
func (i *Interceptor) HandleChallenge(ctx context.Context, challenge *Challenge) (*PaymentPayload, error) {
	if !i.enabled {
		return nil, fmt.Errorf("X402 auto-intercept is disabled")
	}

	// Parse and validate amount
	amount, err := wallet.ParseUSDC(challenge.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid challenge amount: %w", err)
	}

	// Check against auto-pay limit
	if amount.Cmp(i.maxAutoPayAmt) > 0 {
		return nil, fmt.Errorf("X402 payment amount %s exceeds auto-pay limit %s",
			wallet.FormatUSDC(amount), wallet.FormatUSDC(i.maxAutoPayAmt))
	}

	i.logger.Infow("processing X402 payment",
		"url", challenge.PaymentURL,
		"amount", challenge.Amount,
		"recipient", challenge.RecipientAddress,
	)

	// Make the payment
	txHash, err := i.service.HandleX402(ctx, payment.X402Challenge{
		PaymentURL:       challenge.PaymentURL,
		Amount:           challenge.Amount,
		TokenAddress:     challenge.TokenAddress,
		RecipientAddress: challenge.RecipientAddress,
		Network:          challenge.Network,
	})
	if err != nil {
		return nil, fmt.Errorf("x402 payment: %w", err)
	}

	addr, _ := i.service.WalletAddress(ctx)

	return &PaymentPayload{
		TxHash:  txHash,
		From:    addr,
		ChainID: i.service.ChainID(),
	}, nil
}

// IsEnabled returns whether the interceptor is active.
func (i *Interceptor) IsEnabled() bool {
	return i.enabled
}
