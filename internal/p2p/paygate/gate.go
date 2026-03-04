// Package paygate implements a payment gate that checks tool pricing and
// verifies EIP-3009 payment authorizations between the firewall and tool
// executor in the P2P protocol.
package paygate

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/payment/contracts"
	"github.com/langoai/lango/internal/payment/eip3009"

	// wallet is imported for ParseUSDC delegation and currency constants.
	"github.com/langoai/lango/internal/wallet"
)

// PricingFunc returns the price (decimal USDC string like "0.50") and whether
// the tool is free.
type PricingFunc func(toolName string) (price string, isFree bool)

// ResultStatus describes the outcome of a payment gate check.
type ResultStatus string

// DefaultQuoteExpiry is the validity window for a price quote.
const DefaultQuoteExpiry = 5 * time.Minute

const (
	// StatusFree means the tool is free; no payment required.
	StatusFree ResultStatus = "free"

	// StatusVerified means a valid payment authorization was provided.
	StatusVerified ResultStatus = "verified"

	// StatusPaymentRequired means the tool is paid but no authorization was
	// provided; the PriceQuote tells the caller what to pay.
	StatusPaymentRequired ResultStatus = "payment_required"

	// StatusInvalid means the provided payment authorization is invalid.
	StatusInvalid ResultStatus = "invalid"

	// StatusPostPayApproved means the peer has high enough trust to defer
	// payment until after tool execution (post-pay).
	StatusPostPayApproved ResultStatus = "postpay_approved"
)

// Result describes the outcome of a payment gate check.
type Result struct {
	Status       ResultStatus           `json:"status"`
	Auth         *eip3009.Authorization `json:"auth,omitempty"`
	PriceQuote   *PriceQuote            `json:"priceQuote,omitempty"`
	Reason       string                 `json:"reason,omitempty"`
	SettlementID string                 `json:"settlementId,omitempty"`
}

// PriceQuote tells a buyer what to pay for a tool invocation.
type PriceQuote struct {
	ToolName     string `json:"toolName"`
	Price        string `json:"price"`
	Currency     string `json:"currency"`
	USDCContract string `json:"usdcContract"`
	ChainID      int64  `json:"chainId"`
	SellerAddr   string `json:"sellerAddr"`
	QuoteExpiry  int64  `json:"quoteExpiry"`
}

// Config holds construction parameters for a Gate.
type Config struct {
	PricingFn    PricingFunc
	ReputationFn ReputationFunc
	TrustCfg     TrustConfig
	LocalAddr    string
	ChainID      int64
	USDCAddr     common.Address
	RPCClient    *ethclient.Client
	Logger       *zap.SugaredLogger
}

// Gate sits between the firewall and the tool executor, enforcing payment
// requirements for paid tools.
type Gate struct {
	pricingFn    PricingFunc
	reputationFn ReputationFunc
	trustCfg     TrustConfig
	ledger       *DeferredLedger
	localAddr    string
	chainID      int64
	usdcAddr     common.Address
	rpcClient    *ethclient.Client
	logger       *zap.SugaredLogger
}

// New creates a payment gate from the given configuration.
func New(cfg Config) *Gate {
	return &Gate{
		pricingFn:    cfg.PricingFn,
		reputationFn: cfg.ReputationFn,
		trustCfg:     cfg.TrustCfg,
		ledger:       NewDeferredLedger(),
		localAddr:    cfg.LocalAddr,
		chainID:      cfg.ChainID,
		usdcAddr:     cfg.USDCAddr,
		rpcClient:    cfg.RPCClient,
		logger:       cfg.Logger,
	}
}

// Ledger returns the deferred payment ledger for post-pay tracking.
func (g *Gate) Ledger() *DeferredLedger {
	return g.ledger
}

// Check evaluates whether a tool invocation should proceed. It looks up the
// tool price, and if payment is required, validates the EIP-3009 authorization
// embedded in the payload. High-trust peers (score > PostPayMinScore) are
// granted post-pay: the tool executes first, settlement happens asynchronously.
func (g *Gate) Check(peerDID, toolName string, payload map[string]interface{}) (*Result, error) {
	price, isFree := g.pricingFn(toolName)
	if isFree {
		return &Result{Status: StatusFree}, nil
	}

	// Trust-based post-pay: high-reputation peers pay after execution.
	if g.reputationFn != nil {
		score, err := g.reputationFn(context.Background(), peerDID)
		if err != nil {
			g.logger.Warnw("reputation lookup failed, falling back to prepay",
				"peerDID", peerDID, "error", err)
		} else if score > g.trustCfg.PostPayMinScore {
			sid := g.ledger.Add(peerDID, toolName, price)
			g.logger.Infow("post-pay approved",
				"peerDID", peerDID, "tool", toolName, "price", price, "score", score)
			return &Result{
				Status:       StatusPostPayApproved,
				SettlementID: sid,
			}, nil
		}
	}

	// Look for payment authorization in the payload.
	authRaw, ok := payload["paymentAuth"]
	if !ok {
		quote := g.BuildQuote(toolName, price)
		return &Result{
			Status:     StatusPaymentRequired,
			PriceQuote: quote,
		}, nil
	}

	authMap, ok := authRaw.(map[string]interface{})
	if !ok {
		return &Result{
			Status: StatusInvalid,
			Reason: "paymentAuth is not a valid object",
		}, nil
	}

	auth, err := parseAuthorization(authMap)
	if err != nil {
		return &Result{
			Status: StatusInvalid,
			Reason: fmt.Sprintf("parse paymentAuth: %v", err),
		}, nil
	}

	// Verify: recipient must be the local address.
	if auth.To != common.HexToAddress(g.localAddr) {
		return &Result{
			Status: StatusInvalid,
			Reason: fmt.Sprintf("recipient mismatch: got %s, want %s", auth.To.Hex(), g.localAddr),
		}, nil
	}

	// Verify: amount must cover the price.
	requiredAmount, err := ParseUSDC(price)
	if err != nil {
		return nil, fmt.Errorf("parse tool price %q: %w", price, err)
	}
	if auth.Value.Cmp(requiredAmount) < 0 {
		return &Result{
			Status: StatusInvalid,
			Reason: fmt.Sprintf("insufficient payment: got %s, need %s", auth.Value, requiredAmount),
		}, nil
	}

	// Verify: authorization must not be expired.
	now := time.Now().Unix()
	if auth.ValidBefore.Int64() <= now {
		return &Result{
			Status: StatusInvalid,
			Reason: "payment authorization expired",
		}, nil
	}

	// Verify: USDC contract must be canonical for this chain.
	if !contracts.IsCanonical(g.chainID, g.usdcAddr) {
		return &Result{
			Status: StatusInvalid,
			Reason: fmt.Sprintf("non-canonical USDC contract for chain %d", g.chainID),
		}, nil
	}

	return &Result{
		Status: StatusVerified,
		Auth:   auth,
	}, nil
}

// BuildQuote creates a PriceQuote for the given tool and price.
func (g *Gate) BuildQuote(toolName, price string) *PriceQuote {
	return &PriceQuote{
		ToolName:     toolName,
		Price:        price,
		Currency:     wallet.CurrencyUSDC,
		USDCContract: g.usdcAddr.Hex(),
		ChainID:      g.chainID,
		SellerAddr:   g.localAddr,
		QuoteExpiry:  time.Now().Add(DefaultQuoteExpiry).Unix(),
	}
}

// ParseUSDC delegates to wallet.ParseUSDC. Kept as a package-level alias for
// backward compatibility within the paygate package.
var ParseUSDC = wallet.ParseUSDC

// parseAuthorization converts a JSON-decoded map into an eip3009.Authorization.
func parseAuthorization(m map[string]interface{}) (*eip3009.Authorization, error) {
	auth := &eip3009.Authorization{}

	from, err := getHexAddress(m, "from")
	if err != nil {
		return nil, fmt.Errorf("from: %w", err)
	}
	auth.From = from

	to, err := getHexAddress(m, "to")
	if err != nil {
		return nil, fmt.Errorf("to: %w", err)
	}
	auth.To = to

	value, err := getBigInt(m, "value")
	if err != nil {
		return nil, fmt.Errorf("value: %w", err)
	}
	auth.Value = value

	validAfter, err := getBigInt(m, "validAfter")
	if err != nil {
		return nil, fmt.Errorf("validAfter: %w", err)
	}
	auth.ValidAfter = validAfter

	validBefore, err := getBigInt(m, "validBefore")
	if err != nil {
		return nil, fmt.Errorf("validBefore: %w", err)
	}
	auth.ValidBefore = validBefore

	nonce, err := getBytes32(m, "nonce")
	if err != nil {
		return nil, fmt.Errorf("nonce: %w", err)
	}
	auth.Nonce = nonce

	v, err := getUint8(m, "v")
	if err != nil {
		return nil, fmt.Errorf("v: %w", err)
	}
	auth.V = v

	r, err := getBytes32(m, "r")
	if err != nil {
		return nil, fmt.Errorf("r: %w", err)
	}
	auth.R = r

	s, err := getBytes32(m, "s")
	if err != nil {
		return nil, fmt.Errorf("s: %w", err)
	}
	auth.S = s

	return auth, nil
}

// getHexAddress extracts a hex-encoded Ethereum address from a map field.
func getHexAddress(m map[string]interface{}, key string) (common.Address, error) {
	v, ok := m[key]
	if !ok {
		return common.Address{}, fmt.Errorf("missing field %q", key)
	}
	s, ok := v.(string)
	if !ok {
		return common.Address{}, fmt.Errorf("field %q is not a string", key)
	}
	if !common.IsHexAddress(s) {
		return common.Address{}, fmt.Errorf("field %q is not a valid hex address", key)
	}
	return common.HexToAddress(s), nil
}

// getBigInt extracts a big.Int from a map field (accepts string or float64).
func getBigInt(m map[string]interface{}, key string) (*big.Int, error) {
	v, ok := m[key]
	if !ok {
		return nil, fmt.Errorf("missing field %q", key)
	}

	switch val := v.(type) {
	case string:
		n, ok := new(big.Int).SetString(val, 0)
		if !ok {
			return nil, fmt.Errorf("field %q: invalid integer %q", key, val)
		}
		return n, nil
	case float64:
		return big.NewInt(int64(val)), nil
	default:
		return nil, fmt.Errorf("field %q: unsupported type %T", key, v)
	}
}

// getBytes32 extracts a [32]byte from a map field (hex string).
func getBytes32(m map[string]interface{}, key string) ([32]byte, error) {
	var result [32]byte
	v, ok := m[key]
	if !ok {
		return result, fmt.Errorf("missing field %q", key)
	}
	s, ok := v.(string)
	if !ok {
		return result, fmt.Errorf("field %q is not a string", key)
	}
	b := common.FromHex(s)
	if len(b) != 32 {
		return result, fmt.Errorf("field %q: expected 32 bytes, got %d", key, len(b))
	}
	copy(result[:], b)
	return result, nil
}

// getUint8 extracts a uint8 from a map field (float64 from JSON).
func getUint8(m map[string]interface{}, key string) (uint8, error) {
	v, ok := m[key]
	if !ok {
		return 0, fmt.Errorf("missing field %q", key)
	}
	f, ok := v.(float64)
	if !ok {
		return 0, fmt.Errorf("field %q is not a number", key)
	}
	if f < 0 || f > 255 {
		return 0, fmt.Errorf("field %q out of uint8 range: %f", key, f)
	}
	return uint8(f), nil
}
