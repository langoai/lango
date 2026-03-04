package team

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Sentinel errors for payment negotiation.
var (
	ErrPriceRejected   = errors.New("proposed price was rejected")
	ErrNegotiationFail = errors.New("payment negotiation failed")
)

// PaymentMode describes how a team member will be compensated.
type PaymentMode string

const (
	PaymentPrepay  PaymentMode = "prepay"
	PaymentPostpay PaymentMode = "postpay"
	PaymentFree    PaymentMode = "free"
)

// PaymentAgreement records the negotiated payment terms between a team leader and a member.
type PaymentAgreement struct {
	TeamID      string      `json:"teamId"`
	MemberDID   string      `json:"memberDid"`
	Mode        PaymentMode `json:"mode"`
	PricePerUse string      `json:"pricePerUse"` // decimal string, e.g. "0.50"
	Currency    string      `json:"currency"`
	MaxUses     int         `json:"maxUses"`     // 0 = unlimited
	ValidUntil  time.Time   `json:"validUntil"`
	AgreedAt    time.Time   `json:"agreedAt"`
}

// IsExpired reports whether the agreement has passed its validity window.
func (a *PaymentAgreement) IsExpired() bool {
	if a.ValidUntil.IsZero() {
		return false
	}
	return time.Now().After(a.ValidUntil)
}

// DefaultPostPayThreshold is the minimum trust score for post-pay eligibility.
// Matches paygate.DefaultPostPayThreshold to keep both layers consistent.
const DefaultPostPayThreshold = 0.7

// SelectPaymentMode chooses payment mode based on trust score and price.
// High trust (>= DefaultPostPayThreshold) with nonzero price -> PostPay; low trust -> PrePay; zero price -> Free.
func SelectPaymentMode(trustScore, pricePerTask float64) PaymentMode {
	if pricePerTask <= 0 {
		return PaymentFree
	}
	if trustScore >= DefaultPostPayThreshold {
		return PaymentPostpay
	}
	return PaymentPrepay
}

// NegotiatePaymentQuick creates a payment agreement using the simple trust-based mode selection.
func NegotiatePaymentQuick(teamID, agentDID string, trustScore, pricePerTask, maxBudget float64) *PaymentAgreement {
	return &PaymentAgreement{
		TeamID:      teamID,
		MemberDID:   agentDID,
		Mode:        SelectPaymentMode(trustScore, pricePerTask),
		PricePerUse: fmt.Sprintf("%.2f", pricePerTask),
		Currency:    "USDC",
		MaxUses:     int(maxBudget / max(pricePerTask, 0.01)),
		AgreedAt:    time.Now(),
	}
}

// PriceQueryFunc queries a remote agent's price for a capability or tool.
type PriceQueryFunc func(ctx context.Context, peerID, toolName string) (price string, isFree bool, err error)

// TrustScoreFunc retrieves the trust score for a peer.
type TrustScoreFunc func(ctx context.Context, peerDID string) (float64, error)

// NegotiatorConfig configures the payment negotiator.
type NegotiatorConfig struct {
	PriceQueryFn     PriceQueryFunc
	TrustScoreFn     TrustScoreFunc
	PostPayThreshold float64 // min trust score for post-pay (default: 0.8)
	DefaultValidity  time.Duration
}

// Negotiator handles payment negotiation between team leader and members.
type Negotiator struct {
	queryPrice       PriceQueryFunc
	trustScore       TrustScoreFunc
	postPayThreshold float64
	defaultValidity  time.Duration
}

// NewNegotiator creates a payment negotiator.
func NewNegotiator(cfg NegotiatorConfig) *Negotiator {
	threshold := cfg.PostPayThreshold
	if threshold <= 0 {
		threshold = 0.8
	}
	validity := cfg.DefaultValidity
	if validity <= 0 {
		validity = 1 * time.Hour
	}
	return &Negotiator{
		queryPrice:       cfg.PriceQueryFn,
		trustScore:       cfg.TrustScoreFn,
		postPayThreshold: threshold,
		defaultValidity:  validity,
	}
}

// NegotiatePayment determines the payment terms for a team member.
// It queries the member's price and the leader's trust in the member to decide the mode.
func (n *Negotiator) NegotiatePayment(ctx context.Context, teamID string, member *Member, toolName string) (*PaymentAgreement, error) {
	if n.queryPrice == nil {
		// No pricing function — assume free.
		return &PaymentAgreement{
			TeamID:    teamID,
			MemberDID: member.DID,
			Mode:      PaymentFree,
			AgreedAt:  time.Now(),
		}, nil
	}

	price, isFree, err := n.queryPrice(ctx, member.PeerID, toolName)
	if err != nil {
		return nil, fmt.Errorf("query price for %s: %w", member.DID, err)
	}

	if isFree {
		return &PaymentAgreement{
			TeamID:    teamID,
			MemberDID: member.DID,
			Mode:      PaymentFree,
			AgreedAt:  time.Now(),
		}, nil
	}

	// Determine payment mode based on trust.
	mode := PaymentPrepay
	if n.trustScore != nil {
		score, trustErr := n.trustScore(ctx, member.DID)
		if trustErr == nil && score >= n.postPayThreshold {
			mode = PaymentPostpay
		}
	}

	return &PaymentAgreement{
		TeamID:      teamID,
		MemberDID:   member.DID,
		Mode:        mode,
		PricePerUse: price,
		Currency:    "USDC",
		ValidUntil:  time.Now().Add(n.defaultValidity),
		AgreedAt:    time.Now(),
	}, nil
}
