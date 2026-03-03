package team

import (
	"context"
	"testing"
	"time"
)

func TestNegotiatePayment_Free(t *testing.T) {
	n := NewNegotiator(NegotiatorConfig{
		PriceQueryFn: func(_ context.Context, _, _ string) (string, bool, error) {
			return "", true, nil
		},
	})

	member := &Member{DID: "did:1", PeerID: "peer-1"}
	agreement, err := n.NegotiatePayment(context.Background(), "t1", member, "search")
	if err != nil {
		t.Fatalf("NegotiatePayment() error = %v", err)
	}
	if agreement.Mode != PaymentFree {
		t.Errorf("Mode = %q, want %q", agreement.Mode, PaymentFree)
	}
}

func TestNegotiatePayment_Prepay(t *testing.T) {
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
	if err != nil {
		t.Fatalf("NegotiatePayment() error = %v", err)
	}
	if agreement.Mode != PaymentPrepay {
		t.Errorf("Mode = %q, want %q", agreement.Mode, PaymentPrepay)
	}
	if agreement.PricePerUse != "0.50" {
		t.Errorf("PricePerUse = %q, want %q", agreement.PricePerUse, "0.50")
	}
	if agreement.Currency != "USDC" {
		t.Errorf("Currency = %q, want %q", agreement.Currency, "USDC")
	}
}

func TestNegotiatePayment_Postpay(t *testing.T) {
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
	if err != nil {
		t.Fatalf("NegotiatePayment() error = %v", err)
	}
	if agreement.Mode != PaymentPostpay {
		t.Errorf("Mode = %q, want %q", agreement.Mode, PaymentPostpay)
	}
}

func TestNegotiatePayment_NoPriceFunc(t *testing.T) {
	n := NewNegotiator(NegotiatorConfig{})

	member := &Member{DID: "did:1", PeerID: "peer-1"}
	agreement, err := n.NegotiatePayment(context.Background(), "t1", member, "search")
	if err != nil {
		t.Fatalf("NegotiatePayment() error = %v", err)
	}
	if agreement.Mode != PaymentFree {
		t.Errorf("Mode = %q, want %q (no price func means free)", agreement.Mode, PaymentFree)
	}
}

func TestSelectPaymentMode(t *testing.T) {
	tests := []struct {
		give         string
		trustScore   float64
		pricePerTask float64
		want         PaymentMode
	}{
		{give: "free when price is zero", trustScore: 0.9, pricePerTask: 0, want: PaymentFree},
		{give: "free when price is negative", trustScore: 0.5, pricePerTask: -1, want: PaymentFree},
		{give: "postpay for high trust", trustScore: 0.7, pricePerTask: 1.0, want: PaymentPostpay},
		{give: "postpay for very high trust", trustScore: 0.95, pricePerTask: 0.50, want: PaymentPostpay},
		{give: "prepay for low trust", trustScore: 0.5, pricePerTask: 1.0, want: PaymentPrepay},
		{give: "prepay for zero trust", trustScore: 0, pricePerTask: 0.10, want: PaymentPrepay},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := SelectPaymentMode(tt.trustScore, tt.pricePerTask)
			if got != tt.want {
				t.Errorf("SelectPaymentMode(%f, %f) = %q, want %q", tt.trustScore, tt.pricePerTask, got, tt.want)
			}
		})
	}
}

func TestNegotiatePaymentQuick(t *testing.T) {
	a := NegotiatePaymentQuick("t1", "did:1", 0.9, 0.50, 10.0)
	if a.Mode != PaymentPostpay {
		t.Errorf("Mode = %q, want %q", a.Mode, PaymentPostpay)
	}
	if a.PricePerUse != "0.50" {
		t.Errorf("PricePerUse = %q, want %q", a.PricePerUse, "0.50")
	}
	if a.MaxUses != 20 {
		t.Errorf("MaxUses = %d, want 20 (10.0/0.50)", a.MaxUses)
	}
}

func TestPaymentAgreement_IsExpired(t *testing.T) {
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
			a := &PaymentAgreement{ValidUntil: tt.validUntil}
			if got := a.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}
