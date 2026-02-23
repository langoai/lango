package paygate

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// testGate creates a Gate configured for Base Sepolia testnet.
func testGate(pricingFn PricingFunc) *Gate {
	logger := zap.NewNop().Sugar()
	return New(Config{
		PricingFn: pricingFn,
		LocalAddr: "0x1234567890abcdef1234567890abcdef12345678",
		ChainID:   84532, // Base Sepolia
		USDCAddr:  common.HexToAddress("0x036CbD53842c5426634e7929541eC2318f3dCF7e"),
		Logger:    logger,
	})
}

func makeValidAuth(to string, amount *big.Int) map[string]interface{} {
	nonce := "0x0000000000000000000000000000000000000000000000000000000000000001"
	r := "0x0000000000000000000000000000000000000000000000000000000000000002"
	s := "0x0000000000000000000000000000000000000000000000000000000000000003"

	return map[string]interface{}{
		"from":        "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"to":          to,
		"value":       amount.String(),
		"validAfter":  "0",
		"validBefore": fmt.Sprintf("%d", time.Now().Add(10*time.Minute).Unix()),
		"nonce":       nonce,
		"v":           float64(27),
		"r":           r,
		"s":           s,
	}
}

func TestCheck_FreeTool(t *testing.T) {
	gate := testGate(func(toolName string) (string, bool) {
		return "", true
	})

	result, err := gate.Check("did:peer:buyer", "free-tool", nil)
	require.NoError(t, err)
	assert.Equal(t, StatusFree, result.Status)
	assert.Nil(t, result.PriceQuote)
	assert.Nil(t, result.Auth)
}

func TestCheck_PaidNoAuth(t *testing.T) {
	gate := testGate(func(toolName string) (string, bool) {
		return "0.50", false
	})

	result, err := gate.Check("did:peer:buyer", "paid-tool", map[string]interface{}{})
	require.NoError(t, err)
	assert.Equal(t, StatusPaymentRequired, result.Status)
	require.NotNil(t, result.PriceQuote)
	assert.Equal(t, "paid-tool", result.PriceQuote.ToolName)
	assert.Equal(t, "0.50", result.PriceQuote.Price)
	assert.Equal(t, "USDC", result.PriceQuote.Currency)
	assert.Equal(t, "0x1234567890abcdef1234567890abcdef12345678", result.PriceQuote.SellerAddr)
}

func TestCheck_PaidWithValidAuth(t *testing.T) {
	gate := testGate(func(toolName string) (string, bool) {
		return "0.50", false
	})

	amount := big.NewInt(500000) // 0.50 USDC in 6 decimals
	authMap := makeValidAuth("0x1234567890abcdef1234567890abcdef12345678", amount)

	result, err := gate.Check("did:peer:buyer", "paid-tool", map[string]interface{}{
		"paymentAuth": authMap,
	})
	require.NoError(t, err)
	assert.Equal(t, StatusVerified, result.Status)
	require.NotNil(t, result.Auth)
}

func TestCheck_PaidInsufficientAmount(t *testing.T) {
	gate := testGate(func(toolName string) (string, bool) {
		return "1.00", false
	})

	amount := big.NewInt(500000) // 0.50 USDC — insufficient for $1.00
	authMap := makeValidAuth("0x1234567890abcdef1234567890abcdef12345678", amount)

	result, err := gate.Check("did:peer:buyer", "paid-tool", map[string]interface{}{
		"paymentAuth": authMap,
	})
	require.NoError(t, err)
	assert.Equal(t, StatusInvalid, result.Status)
	assert.Contains(t, result.Reason, "insufficient payment")
}

func TestCheck_ExpiredAuth(t *testing.T) {
	gate := testGate(func(toolName string) (string, bool) {
		return "0.50", false
	})

	amount := big.NewInt(500000)
	authMap := makeValidAuth("0x1234567890abcdef1234567890abcdef12345678", amount)
	// Set validBefore to the past.
	authMap["validBefore"] = fmt.Sprintf("%d", time.Now().Add(-10*time.Minute).Unix())

	result, err := gate.Check("did:peer:buyer", "paid-tool", map[string]interface{}{
		"paymentAuth": authMap,
	})
	require.NoError(t, err)
	assert.Equal(t, StatusInvalid, result.Status)
	assert.Contains(t, result.Reason, "expired")
}

func TestCheck_RecipientMismatch(t *testing.T) {
	gate := testGate(func(toolName string) (string, bool) {
		return "0.50", false
	})

	amount := big.NewInt(500000)
	// Wrong recipient address.
	authMap := makeValidAuth("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef", amount)

	result, err := gate.Check("did:peer:buyer", "paid-tool", map[string]interface{}{
		"paymentAuth": authMap,
	})
	require.NoError(t, err)
	assert.Equal(t, StatusInvalid, result.Status)
	assert.Contains(t, result.Reason, "recipient mismatch")
}

func TestCheck_InvalidAuthType(t *testing.T) {
	gate := testGate(func(toolName string) (string, bool) {
		return "0.50", false
	})

	result, err := gate.Check("did:peer:buyer", "paid-tool", map[string]interface{}{
		"paymentAuth": "not-a-map",
	})
	require.NoError(t, err)
	assert.Equal(t, StatusInvalid, result.Status)
	assert.Contains(t, result.Reason, "not a valid object")
}

func TestParseUSDC(t *testing.T) {
	tests := []struct {
		give    string
		want    int64
		wantErr bool
	}{
		{give: "0.50", want: 500000},
		{give: "1.00", want: 1000000},
		{give: "0.000001", want: 1},
		{give: "100", want: 100000000},
		{give: "0", want: 0},
		{give: "1.123456", want: 1123456},
		{give: "0.0000001", wantErr: true},
		{give: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got, err := ParseUSDC(tt.give)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, big.NewInt(tt.want), got)
		})
	}
}

func TestBuildQuote(t *testing.T) {
	gate := testGate(nil)
	quote := gate.BuildQuote("my-tool", "2.50")

	assert.Equal(t, "my-tool", quote.ToolName)
	assert.Equal(t, "2.50", quote.Price)
	assert.Equal(t, "USDC", quote.Currency)
	assert.Equal(t, int64(84532), quote.ChainID)
	assert.Equal(t, "0x1234567890abcdef1234567890abcdef12345678", quote.SellerAddr)
	assert.Greater(t, quote.QuoteExpiry, time.Now().Unix())
}
