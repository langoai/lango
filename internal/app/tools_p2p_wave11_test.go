package app

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/p2p/firewall"
	"github.com/langoai/lango/internal/p2p/handshake"
	"github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/p2p/protocol"
	"github.com/langoai/lango/internal/payment/eip3009"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestP2PToolsWave11_InventoryAndSimpleHandlers(t *testing.T) {
	t.Parallel()

	sessions, err := handshake.NewSessionStore(time.Hour)
	require.NoError(t, err)
	_, err = sessions.Create("did:lango:peer-wave11", true)
	require.NoError(t, err)

	fw := firewall.New(nil, zap.NewNop().Sugar())
	pc := &p2pComponents{sessions: sessions, fw: fw}
	tools := buildP2PTools(pc)
	names := make(map[string]agent.SafetyLevel, len(tools))
	for _, tool := range tools {
		names[tool.Name] = tool.SafetyLevel
	}
	for _, name := range []string{
		"p2p_status",
		"p2p_connect",
		"p2p_disconnect",
		"p2p_peers",
		"p2p_query",
		"p2p_firewall_rules",
		"p2p_firewall_add",
		"p2p_firewall_remove",
		"p2p_price_query",
		"p2p_reputation",
		"p2p_discover",
	} {
		assert.Contains(t, names, name)
	}
	assert.Equal(t, agent.SafetyLevelSafe, names["p2p_peers"])
	assert.Equal(t, agent.SafetyLevelDangerous, names["p2p_firewall_add"])

	result, err := findP2PTool(t, tools, "p2p_peers").Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	peers := wave11Payload(t, result)
	assert.Equal(t, 1, peers["count"])

	result, err = findP2PTool(t, tools, "p2p_firewall_add").Handler(context.Background(), map[string]interface{}{
		"peer_did":   "did:lango:peer-wave11",
		"action":     "allow",
		"tools":      []interface{}{"search_knowledge", "payment_balance"},
		"rate_limit": float64(3),
	})
	require.NoError(t, err)
	assert.Equal(t, "added", wave11Payload(t, result)["status"])

	result, err = findP2PTool(t, tools, "p2p_firewall_rules").Handler(context.Background(), map[string]interface{}{})
	require.NoError(t, err)
	rules := wave11Payload(t, result)
	assert.Equal(t, 1, rules["count"])

	result, err = findP2PTool(t, tools, "p2p_firewall_remove").Handler(context.Background(), map[string]interface{}{
		"peer_did": "did:lango:peer-wave11",
	})
	require.NoError(t, err)
	assert.Equal(t, 1, wave11Payload(t, result)["count"])

	result, err = findP2PTool(t, tools, "p2p_discover").Handler(context.Background(), map[string]interface{}{"capability": "search"})
	require.NoError(t, err)
	discovery := wave11Payload(t, result)
	assert.Equal(t, 0, discovery["count"])
	assert.Equal(t, "gossip not enabled", discovery["message"])

	_, err = findP2PTool(t, tools, "p2p_query").Handler(context.Background(), map[string]interface{}{
		"peer_did":  "did:lango:missing",
		"tool_name": "search_knowledge",
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "no active session")

	_, err = findP2PTool(t, tools, "p2p_price_query").Handler(context.Background(), map[string]interface{}{
		"peer_did":  "did:lango:missing",
		"tool_name": "search_knowledge",
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "connect first")

	_, err = findP2PTool(t, tools, "p2p_reputation").Handler(context.Background(), map[string]interface{}{
		"peer_did": "did:lango:peer-wave11",
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "reputation system not available")

	result, err = findP2PTool(t, tools, "p2p_disconnect").Handler(context.Background(), map[string]interface{}{
		"peer_did": "did:lango:peer-wave11",
	})
	require.NoError(t, err)
	assert.Equal(t, "disconnected", wave11Payload(t, result)["status"])
	assert.Nil(t, sessions.Get("did:lango:peer-wave11"))
}

func TestAuthToMapWave11_SerializesAuthorizationFields(t *testing.T) {
	t.Parallel()

	auth := &eip3009.Authorization{
		From:        common.HexToAddress("0x1111111111111111111111111111111111111111"),
		To:          common.HexToAddress("0x2222222222222222222222222222222222222222"),
		Value:       big.NewInt(500000),
		ValidAfter:  big.NewInt(12),
		ValidBefore: big.NewInt(34),
		Nonce:       [32]byte{31: 1},
		V:           28,
		R:           [32]byte{31: 2},
		S:           [32]byte{31: 3},
	}

	got := authToMap(auth)

	assert.Equal(t, "0x1111111111111111111111111111111111111111", got["from"])
	assert.Equal(t, "0x2222222222222222222222222222222222222222", got["to"])
	assert.Equal(t, "500000", got["value"])
	assert.Equal(t, "12", got["validAfter"])
	assert.Equal(t, "34", got["validBefore"])
	assert.Equal(t, float64(28), got["v"])
	assert.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000000001", got["nonce"])
	assert.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000000002", got["r"])
	assert.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000000003", got["s"])
}

func TestP2PPaidInvokeWave11_FreeAndPaymentBranches(t *testing.T) {
	originalRemoteAgent := newP2PPaidInvokeRemoteAgent
	t.Cleanup(func() {
		newP2PPaidInvokeRemoteAgent = originalRemoteAgent
	})

	did := wave11PeerDID(t)
	sessions, err := handshake.NewSessionStore(time.Hour)
	require.NoError(t, err)
	_, err = sessions.Create(did.ID, false)
	require.NoError(t, err)

	remote := &wave11PaidInvokeRemoteAgent{
		quote: &protocol.PriceQuoteResult{
			ToolName: "search_knowledge",
			IsFree:   true,
		},
		invokeResult: map[string]interface{}{"answer": "free-result"},
	}
	newP2PPaidInvokeRemoteAgent = func(string, *identity.DID, *handshake.Session, *p2pComponents) p2pPaidInvokeRemoteAgent {
		return remote
	}
	tools := buildP2PPaidInvokeTool(&p2pComponents{sessions: sessions}, &paymentComponents{
		wallet:  newP2PTestWallet(t),
		limiter: p2pTestLimiter{},
		chainID: 84532,
	})
	require.Len(t, tools, 1)

	result, err := tools[0].Handler(context.Background(), map[string]interface{}{
		"peer_did":  did.ID,
		"tool_name": "search_knowledge",
		"params":    `{"query":"status"}`,
	})
	require.NoError(t, err)
	payload := wave11Payload(t, result)
	assert.Equal(t, "ok", payload["status"])
	assert.Equal(t, false, payload["paid"])
	assert.Equal(t, map[string]interface{}{"answer": "free-result"}, payload["result"])
	assert.Equal(t, map[string]interface{}{"query": "status"}, remote.invokeParams)

	remote.quote = &protocol.PriceQuoteResult{
		ToolName:   "paid_tool",
		Price:      "0.75",
		Currency:   "USDC",
		SellerAddr: "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		ChainID:    84532,
		IsFree:     false,
	}
	result, err = tools[0].Handler(context.Background(), map[string]interface{}{
		"peer_did":  did.ID,
		"tool_name": "paid_tool",
	})
	require.NoError(t, err)
	payload = wave11Payload(t, result)
	assert.Equal(t, "approval_required", payload["status"])
	assert.Equal(t, "0.75", payload["price"])
	assert.Equal(t, "USDC", payload["currency"])
	assert.False(t, remote.paidInvoked)
}

func TestP2PPaidInvokeWave11_PaidResponseBranches(t *testing.T) {
	originalRemoteAgent := newP2PPaidInvokeRemoteAgent
	t.Cleanup(func() {
		newP2PPaidInvokeRemoteAgent = originalRemoteAgent
	})

	did := wave11PeerDID(t)
	sessions, err := handshake.NewSessionStore(time.Hour)
	require.NoError(t, err)
	_, err = sessions.Create(did.ID, false)
	require.NoError(t, err)

	remote := &wave11PaidInvokeRemoteAgent{
		quote: &protocol.PriceQuoteResult{
			ToolName:   "paid_tool",
			Price:      "0.25",
			Currency:   "USDC",
			SellerAddr: "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			ChainID:    84532,
			IsFree:     false,
		},
		paidResponse: &protocol.Response{Status: protocol.ResponseStatusPaymentRequired, Result: map[string]interface{}{"reason": "expired"}},
	}
	newP2PPaidInvokeRemoteAgent = func(string, *identity.DID, *handshake.Session, *p2pComponents) p2pPaidInvokeRemoteAgent {
		return remote
	}
	tools := buildP2PPaidInvokeTool(&p2pComponents{sessions: sessions}, &paymentComponents{
		wallet:  newP2PTestWallet(t),
		limiter: p2pAutoApproveLimiter{},
		chainID: 84532,
	})
	require.Len(t, tools, 1)

	result, err := tools[0].Handler(context.Background(), map[string]interface{}{
		"peer_did":  did.ID,
		"tool_name": "paid_tool",
	})
	require.NoError(t, err)
	payload := wave11Payload(t, result)
	assert.Equal(t, "payment_required", payload["status"])
	assert.Equal(t, map[string]interface{}{"reason": "expired"}, payload["detail"])
	require.True(t, remote.paidInvoked)
	assert.Equal(t, "250000", remote.lastAuth["value"])

	remote.paidResponse = &protocol.Response{Status: protocol.ResponseStatusError, Error: "seller failed"}
	_, err = tools[0].Handler(context.Background(), map[string]interface{}{
		"peer_did":  did.ID,
		"tool_name": "paid_tool",
	})
	require.Error(t, err)
	assert.ErrorContains(t, err, "remote paid_tool: seller failed")
}

type wave11PaidInvokeRemoteAgent struct {
	quote        *protocol.PriceQuoteResult
	invokeResult map[string]interface{}
	invokeParams map[string]interface{}
	paidResponse *protocol.Response
	paidInvoked  bool
	lastAuth     map[string]interface{}
}

func (a *wave11PaidInvokeRemoteAgent) QueryPrice(context.Context, string) (*protocol.PriceQuoteResult, error) {
	return a.quote, nil
}

func (a *wave11PaidInvokeRemoteAgent) InvokeTool(_ context.Context, _ string, params map[string]interface{}) (map[string]interface{}, error) {
	a.invokeParams = params
	return a.invokeResult, nil
}

func (a *wave11PaidInvokeRemoteAgent) InvokeToolPaid(_ context.Context, _ string, _ map[string]interface{}, auth map[string]interface{}) (*protocol.Response, error) {
	a.paidInvoked = true
	a.lastAuth = auth
	return a.paidResponse, nil
}

func wave11PeerDID(t *testing.T) *identity.DID {
	t.Helper()

	pk, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	did, err := identity.DIDFromPublicKey(ethcrypto.CompressPubkey(&pk.PublicKey))
	require.NoError(t, err)
	return did
}

func wave11Payload(t *testing.T, result interface{}) map[string]interface{} {
	t.Helper()

	payload, ok := result.(map[string]interface{})
	require.True(t, ok)
	return payload
}
