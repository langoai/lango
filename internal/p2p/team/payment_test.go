package team

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNegotiatePayment_Free(t *testing.T) {
	t.Parallel()

	n := NewNegotiator(NegotiatorConfig{
		PriceQueryFn: func(_ context.Context, _, _ string) (string, bool, error) {
			return "", true, nil
		},
	})

	member := &Member{DID: "did:1", PeerID: "peer-1"}
	agreement, err := n.NegotiatePayment(context.Background(), "t1", member, "search")
	require.NoError(t, err)
	assert.Equal(t, PaymentFree, agreement.Mode)
}

func TestNegotiatePayment_Prepay(t *testing.T) {
	t.Parallel()

	n := NewNegotiator(NegotiatorConfig{
		PriceQueryFn: func(_ context.Context, _, _ string) (string, bool, error) {
			return "0.50", false, nil
		},
		TrustScoreFn: func(_ context.Context, _ string) (float64, error) {
			return 0.5, nil // below threshold
		},
		PostPayThreshold: 0.8,
	})

	member := &Member{DID: "did:1", PeerID: "peer-1"}
	agreement, err := n.NegotiatePayment(context.Background(), "t1", member, "search")
	require.NoError(t, err)
	assert.Equal(t, PaymentPrepay, agreement.Mode)
	assert.Equal(t, "0.50", agreement.PricePerUse)
	assert.Equal(t, "USDC", agreement.Currency)
}

func TestNegotiatePayment_Postpay(t *testing.T) {
	t.Parallel()

	n := NewNegotiator(NegotiatorConfig{
		PriceQueryFn: func(_ context.Context, _, _ string) (string, bool, error) {
			return "1.00", false, nil
		},
		TrustScoreFn: func(_ context.Context, _ string) (float64, error) {
			return 0.95, nil // above threshold
		},
		PostPayThreshold: 0.8,
	})

	member := &Member{DID: "did:1", PeerID: "peer-1"}
	agreement, err := n.NegotiatePayment(context.Background(), "t1", member, "search")
	require.NoError(t, err)
	assert.Equal(t, PaymentPostpay, agreement.Mode)
}

func TestNegotiatePayment_NoPriceFunc(t *testing.T) {
	t.Parallel()

	n := NewNegotiator(NegotiatorConfig{})

	member := &Member{DID: "did:1", PeerID: "peer-1"}
	agreement, err := n.NegotiatePayment(context.Background(), "t1", member, "search")
	require.NoError(t, err)
	assert.Equal(t, PaymentFree, agreement.Mode, "no price func means free")
}

func TestSelectPaymentMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give         string
		trustScore   float64
		pricePerTask float64
		want         PaymentMode
	}{
		{give: "free when price is zero", trustScore: 0.9, pricePerTask: 0, want: PaymentFree},
		{give: "free when price is negative", trustScore: 0.5, pricePerTask: -1, want: PaymentFree},
		{give: "postpay for threshold trust", trustScore: 0.8, pricePerTask: 1.0, want: PaymentPostpay},
		{give: "postpay for very high trust", trustScore: 0.95, pricePerTask: 0.50, want: PaymentPostpay},
		{give: "prepay for low trust", trustScore: 0.5, pricePerTask: 1.0, want: PaymentPrepay},
		{give: "prepay for zero trust", trustScore: 0, pricePerTask: 0.10, want: PaymentPrepay},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			got := SelectPaymentMode(tt.trustScore, tt.pricePerTask)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNegotiatePaymentQuick(t *testing.T) {
	t.Parallel()

	a := NegotiatePaymentQuick("t1", "did:1", 0.9, 0.50, 10.0)
	assert.Equal(t, PaymentPostpay, a.Mode)
	assert.Equal(t, "0.50", a.PricePerUse)
	assert.Equal(t, 20, a.MaxUses, "10.0/0.50")
}

func TestPaymentAgreement_IsExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give       string
		validUntil time.Time
		want       bool
	}{
		{give: "zero value (never expires)", validUntil: time.Time{}, want: false},
		{give: "future", validUntil: time.Now().Add(1 * time.Hour), want: false},
		{give: "past", validUntil: time.Now().Add(-1 * time.Hour), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()
			a := &PaymentAgreement{ValidUntil: tt.validUntil}
			assert.Equal(t, tt.want, a.IsExpired())
		})
	}
}
