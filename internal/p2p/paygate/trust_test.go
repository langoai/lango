package paygate

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func testGateWithReputation(pricingFn PricingFunc, repFn ReputationFunc, trustCfg TrustConfig) *Gate {
	logger := zap.NewNop().Sugar()
	return New(Config{
		PricingFn:    pricingFn,
		ReputationFn: repFn,
		TrustCfg:     trustCfg,
		LocalAddr:    "0x1234567890abcdef1234567890abcdef12345678",
		ChainID:      84532,
		Logger:       logger,
	})
}

func paidPricingFn(toolName string) (string, bool) {
	return "0.50", false
}

func TestCheck_HighTrust_PostPay(t *testing.T) {
	repFn := func(ctx context.Context, peerDID string) (float64, error) {
		return 0.9, nil
	}
	gate := testGateWithReputation(paidPricingFn, repFn, DefaultTrustConfig())

	result, err := gate.Check("did:peer:trusted", "paid-tool", nil)
	require.NoError(t, err)
	assert.Equal(t, StatusPostPayApproved, result.Status)
	assert.NotEmpty(t, result.SettlementID)
	assert.Nil(t, result.Auth)
	assert.Nil(t, result.PriceQuote)
}

func TestCheck_MediumTrust_Prepay(t *testing.T) {
	repFn := func(ctx context.Context, peerDID string) (float64, error) {
		return 0.5, nil
	}
	gate := testGateWithReputation(paidPricingFn, repFn, DefaultTrustConfig())

	result, err := gate.Check("did:peer:medium", "paid-tool", map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, StatusPaymentRequired, result.Status)
	assert.Empty(t, result.SettlementID)
}

func TestCheck_ExactThreshold_Prepay(t *testing.T) {
	repFn := func(ctx context.Context, peerDID string) (float64, error) {
		return DefaultPostPayThreshold, nil // exactly at threshold — NOT post-pay (must be strictly greater)
	}
	gate := testGateWithReputation(paidPricingFn, repFn, DefaultTrustConfig())

	result, err := gate.Check("did:peer:borderline", "paid-tool", map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, StatusPaymentRequired, result.Status)
}

func TestCheck_NilReputation_Prepay(t *testing.T) {
	gate := testGateWithReputation(paidPricingFn, nil, DefaultTrustConfig())

	result, err := gate.Check("did:peer:unknown", "paid-tool", map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, StatusPaymentRequired, result.Status)
}

func TestCheck_ReputationError_FallbackPrepay(t *testing.T) {
	repFn := func(ctx context.Context, peerDID string) (float64, error) {
		return 0, errors.New("db unavailable")
	}
	gate := testGateWithReputation(paidPricingFn, repFn, DefaultTrustConfig())

	result, err := gate.Check("did:peer:error", "paid-tool", map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, StatusPaymentRequired, result.Status)
}

func TestCheck_FreeTool_IgnoresReputation(t *testing.T) {
	repFn := func(ctx context.Context, peerDID string) (float64, error) {
		return 1.0, nil
	}
	freeFn := func(toolName string) (string, bool) { return "", true }
	gate := testGateWithReputation(freeFn, repFn, DefaultTrustConfig())

	result, err := gate.Check("did:peer:trusted", "free-tool", nil)
	require.NoError(t, err)
	assert.Equal(t, StatusFree, result.Status)
}

func TestCheck_HighTrust_WithAuth_StillPostPay(t *testing.T) {
	// If peer has high trust AND provides auth, post-pay should take priority
	// (auth is ignored since they qualify for post-pay).
	repFn := func(ctx context.Context, peerDID string) (float64, error) {
		return 0.95, nil
	}
	gate := testGateWithReputation(paidPricingFn, repFn, DefaultTrustConfig())

	amount := big.NewInt(500000)
	authMap := makeValidAuth("0x1234567890abcdef1234567890abcdef12345678", amount)

	result, err := gate.Check("did:peer:trusted", "paid-tool", map[string]interface{}{
		"paymentAuth": authMap,
	})
	require.NoError(t, err)
	assert.Equal(t, StatusPostPayApproved, result.Status)
	assert.NotEmpty(t, result.SettlementID)
}

func TestCheck_CustomThreshold(t *testing.T) {
	repFn := func(ctx context.Context, peerDID string) (float64, error) {
		return 0.7, nil
	}
	cfg := TrustConfig{PostPayMinScore: 0.6}
	gate := testGateWithReputation(paidPricingFn, repFn, cfg)

	result, err := gate.Check("did:peer:custom", "paid-tool", nil)
	require.NoError(t, err)
	assert.Equal(t, StatusPostPayApproved, result.Status)
}
