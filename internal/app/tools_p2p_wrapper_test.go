package app

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/p2p/handshake"
	"github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/p2p/protocol"
	"github.com/langoai/lango/internal/payment/eip3009"
)

func TestBuildP2PTools_RequireCanonicalInputs(t *testing.T) {
	t.Parallel()

	tools := buildP2PTools(&p2pComponents{})

	testCases := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{name: "connect requires multiaddr", tool: "p2p_connect", params: map[string]interface{}{}, wantErr: "missing multiaddr parameter"},
		{name: "disconnect requires peer did", tool: "p2p_disconnect", params: map[string]interface{}{}, wantErr: "missing peer_did parameter"},
		{name: "query requires peer did", tool: "p2p_query", params: map[string]interface{}{"tool_name": "search_knowledge"}, wantErr: "missing peer_did parameter"},
		{name: "query requires tool name", tool: "p2p_query", params: map[string]interface{}{"peer_did": "did:lango:abc"}, wantErr: "missing tool_name parameter"},
		{name: "firewall add requires peer did", tool: "p2p_firewall_add", params: map[string]interface{}{"action": "allow"}, wantErr: "missing peer_did parameter"},
		{name: "firewall add requires action", tool: "p2p_firewall_add", params: map[string]interface{}{"peer_did": "did:lango:abc"}, wantErr: "missing action parameter"},
		{name: "firewall remove requires peer did", tool: "p2p_firewall_remove", params: map[string]interface{}{}, wantErr: "missing peer_did parameter"},
		{name: "price query requires peer did", tool: "p2p_price_query", params: map[string]interface{}{"tool_name": "search_knowledge"}, wantErr: "missing peer_did parameter"},
		{name: "price query requires tool name", tool: "p2p_price_query", params: map[string]interface{}{"peer_did": "did:lango:abc"}, wantErr: "missing tool_name parameter"},
		{name: "reputation requires peer did", tool: "p2p_reputation", params: map[string]interface{}{}, wantErr: "missing peer_did parameter"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := findP2PTool(t, tools, tc.tool).Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func TestBuildP2PPaidInvokeTool_RequireCanonicalInputs(t *testing.T) {
	t.Parallel()

	pc := &paymentComponents{
		wallet:  newP2PTestWallet(t),
		limiter: p2pTestLimiter{},
		chainID: 84532,
	}
	tools := buildP2PPaidInvokeTool(&p2pComponents{}, pc)
	require.Len(t, tools, 1)

	testCases := []struct {
		name    string
		params  map[string]interface{}
		wantErr string
	}{
		{name: "invoke paid requires peer did", params: map[string]interface{}{"tool_name": "search_knowledge"}, wantErr: "missing peer_did parameter"},
		{name: "invoke paid requires tool name", params: map[string]interface{}{"peer_did": "did:lango:abc"}, wantErr: "missing tool_name parameter"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tools[0].Handler(context.Background(), tc.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.EqualError(t, err, tc.wantErr)
		})
	}
}

func TestBuildP2PPaidInvokeTool_StopsBeforeSigningWhenAuthorizationCreationFails(t *testing.T) {
	originalRemoteAgent := newP2PPaidInvokeRemoteAgent
	originalNewUnsigned := newEIP3009Unsigned
	t.Cleanup(func() {
		newP2PPaidInvokeRemoteAgent = originalRemoteAgent
		newEIP3009Unsigned = originalNewUnsigned
	})

	pk, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	did, err := identity.DIDFromPublicKey(ethcrypto.CompressPubkey(&pk.PublicKey))
	require.NoError(t, err)

	sessions, err := handshake.NewSessionStore(time.Hour)
	require.NoError(t, err)
	_, err = sessions.Create(did.ID, false)
	require.NoError(t, err)

	remote := &fakePaidInvokeRemoteAgent{
		quote: &protocol.PriceQuoteResult{
			ToolName:   "search_knowledge",
			Price:      "0.50",
			Currency:   "USDC",
			SellerAddr: "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			ChainID:    84532,
			IsFree:     false,
		},
	}
	newP2PPaidInvokeRemoteAgent = func(string, *identity.DID, *handshake.Session, *p2pComponents) p2pPaidInvokeRemoteAgent {
		return remote
	}
	newEIP3009Unsigned = func(common.Address, common.Address, *big.Int, time.Time) (*eip3009.UnsignedAuth, error) {
		return nil, errors.New("entropy unavailable")
	}

	wallet := newP2PTestWallet(t)
	pc := &paymentComponents{
		wallet:  wallet,
		limiter: p2pAutoApproveLimiter{},
		chainID: 84532,
	}
	tools := buildP2PPaidInvokeTool(&p2pComponents{sessions: sessions}, pc)
	require.Len(t, tools, 1)

	got, err := tools[0].Handler(context.Background(), map[string]interface{}{
		"peer_did":  did.ID,
		"tool_name": "search_knowledge",
	})

	require.Error(t, err)
	assert.Nil(t, got)
	assert.ErrorContains(t, err, "create EIP-3009 authorization")
	assert.ErrorContains(t, err, "entropy unavailable")
	assert.Equal(t, 0, wallet.signTransactionCalls)
	assert.False(t, remote.paidInvoked)
}

type p2pAutoApproveLimiter struct {
	p2pTestLimiter
}

func (p2pAutoApproveLimiter) IsAutoApprovable(context.Context, *big.Int) (bool, error) {
	return true, nil
}

type fakePaidInvokeRemoteAgent struct {
	quote       *protocol.PriceQuoteResult
	paidInvoked bool
}

func (f *fakePaidInvokeRemoteAgent) QueryPrice(context.Context, string) (*protocol.PriceQuoteResult, error) {
	return f.quote, nil
}

func (f *fakePaidInvokeRemoteAgent) InvokeTool(context.Context, string, map[string]interface{}) (map[string]interface{}, error) {
	return nil, errors.New("free invocation should not be called")
}

func (f *fakePaidInvokeRemoteAgent) InvokeToolPaid(context.Context, string, map[string]interface{}, map[string]interface{}) (*protocol.Response, error) {
	f.paidInvoked = true
	return &protocol.Response{Status: protocol.ResponseStatusOK}, nil
}

func findP2PTool(t *testing.T, tools []*agent.Tool, name string) *agent.Tool {
	t.Helper()

	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}

	t.Fatalf("tool %q not found", name)
	return nil
}
