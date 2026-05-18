package app

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"strings"
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

func TestWave28P2PTools_MetadataAndMissingDependencyBranches(t *testing.T) {
	t.Parallel()

	tools := buildP2PTools(&p2pComponents{})
	require.Len(t, tools, 11)

	queryTool := findP2PTool(t, tools, "p2p_query")
	assert.Equal(t, agent.SafetyLevelModerate, queryTool.SafetyLevel)
	assert.Equal(t, "p2p", queryTool.Capability.Category)
	assert.Equal(t, agent.ActivityExecute, queryTool.Capability.Activity)
	assert.False(t, queryTool.Capability.ReadOnly)
	assert.Equal(t, []string{"peer_did", "tool_name"}, wave28P2PRequiredParams(t, queryTool))
	assert.Equal(t, "string", wave28P2PParamType(t, queryTool, "params"))

	priceTool := findP2PTool(t, tools, "p2p_price_query")
	assert.Equal(t, agent.SafetyLevelSafe, priceTool.SafetyLevel)
	assert.True(t, priceTool.Capability.ReadOnly)
	assert.True(t, priceTool.Capability.ConcurrencySafe)
	assert.Equal(t, []string{"peer_did", "tool_name"}, wave28P2PRequiredParams(t, priceTool))

	firewallTool := findP2PTool(t, tools, "p2p_firewall_add")
	assert.Equal(t, agent.SafetyLevelDangerous, firewallTool.SafetyLevel)
	assert.Equal(t, agent.ActivityManage, firewallTool.Capability.Activity)
	assert.Equal(t, []string{"peer_did", "action"}, wave28P2PRequiredParams(t, firewallTool))

	validPayment := &paymentComponents{
		wallet:  wave28P2PWallet(t),
		limiter: &wave28P2PLimiter{},
		chainID: 84532,
	}
	paidTools := buildP2PPaidInvokeTool(&p2pComponents{}, validPayment)
	require.Len(t, paidTools, 1)
	assert.Equal(t, "p2p_invoke_paid", paidTools[0].Name)
	assert.Equal(t, agent.SafetyLevelDangerous, paidTools[0].SafetyLevel)
	assert.Equal(t, agent.ActivityExecute, paidTools[0].Capability.Activity)
	assert.Equal(t, []string{"peer_did", "tool_name"}, wave28P2PRequiredParams(t, paidTools[0]))

	assert.Nil(t, buildP2PPaidInvokeTool(&p2pComponents{}, nil))
	assert.Nil(t, buildP2PPaidInvokeTool(&p2pComponents{}, &paymentComponents{
		limiter: &wave28P2PLimiter{},
		chainID: 84532,
	}))
	assert.Nil(t, buildP2PPaidInvokeTool(&p2pComponents{}, &paymentComponents{
		wallet:  wave28P2PWallet(t),
		chainID: 84532,
	}))
	assert.Nil(t, buildP2PPaidInvokeTool(&p2pComponents{}, &paymentComponents{
		wallet:  wave28P2PWallet(t),
		limiter: &wave28P2PLimiter{},
		chainID: 31337,
	}))
}

func TestP2PToolsWave28_ValidationBranchesAvoidNetwork(t *testing.T) {
	t.Parallel()

	validDID := wave28P2PPeerDID(t)
	sessions := wave28P2PSessions(t, validDID.ID, "not-a-did")
	pc := &p2pComponents{
		sessions: sessions,
		fw:       firewall.New(nil, zap.NewNop().Sugar()),
	}
	tools := buildP2PTools(pc)

	tests := []struct {
		name    string
		tool    string
		params  map[string]interface{}
		wantErr string
	}{
		{
			name:    "connect rejects malformed multiaddr before node access",
			tool:    "p2p_connect",
			params:  map[string]interface{}{"multiaddr": "not-a-multiaddr"},
			wantErr: "invalid multiaddr",
		},
		{
			name: "query rejects active malformed peer DID before remote agent construction",
			tool: "p2p_query",
			params: map[string]interface{}{
				"peer_did":  "not-a-did",
				"tool_name": "search_knowledge",
			},
			wantErr: "parse peer DID",
		},
		{
			name: "query rejects malformed JSON params before remote agent construction",
			tool: "p2p_query",
			params: map[string]interface{}{
				"peer_did":  validDID.ID,
				"tool_name": "search_knowledge",
				"params":    `{"query":`,
			},
			wantErr: "parse params JSON",
		},
		{
			name: "price query rejects active malformed peer DID before remote agent construction",
			tool: "p2p_price_query",
			params: map[string]interface{}{
				"peer_did":  "not-a-did",
				"tool_name": "search_knowledge",
			},
			wantErr: "parse peer DID",
		},
		{
			name: "firewall add surfaces overly permissive validation error",
			tool: "p2p_firewall_add",
			params: map[string]interface{}{
				"peer_did": "*",
				"action":   "allow",
				"tools":    []interface{}{"*"},
			},
			wantErr: "add firewall rule: overly permissive rule",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := findP2PTool(t, tools, tt.tool).Handler(context.Background(), tt.params)
			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestP2PPaidInvokeWave28_ErrorBranchesAvoidNetwork(t *testing.T) {
	originalRemoteAgent := newP2PPaidInvokeRemoteAgent
	t.Cleanup(func() {
		newP2PPaidInvokeRemoteAgent = originalRemoteAgent
	})

	validDID := wave28P2PPeerDID(t)
	sessions := wave28P2PSessions(t, validDID.ID, "not-a-did")
	remote := &wave28P2PRemoteAgent{}
	newP2PPaidInvokeRemoteAgent = func(string, *identity.DID, *handshake.Session, *p2pComponents) p2pPaidInvokeRemoteAgent {
		return remote
	}

	tests := []struct {
		name        string
		params      map[string]interface{}
		quote       *protocol.PriceQuoteResult
		quoteErr    error
		invokeErr   error
		limiter     *wave28P2PLimiter
		wallet      *wave28P2PTestWallet
		wantErr     string
		wantQueries int
	}{
		{
			name: "missing session stops before DID parsing and remote construction",
			params: map[string]interface{}{
				"peer_did":  "did:lango:missing",
				"tool_name": "search_knowledge",
			},
			limiter: &wave28P2PLimiter{autoOK: true},
			wallet:  wave28P2PWallet(t),
			wantErr: "no active session",
		},
		{
			name: "malformed active DID stops before remote construction",
			params: map[string]interface{}{
				"peer_did":  "not-a-did",
				"tool_name": "search_knowledge",
			},
			limiter: &wave28P2PLimiter{autoOK: true},
			wallet:  wave28P2PWallet(t),
			wantErr: "parse peer DID",
		},
		{
			name: "malformed params JSON stops before remote construction",
			params: map[string]interface{}{
				"peer_did":  validDID.ID,
				"tool_name": "search_knowledge",
				"params":    `{"query":`,
			},
			limiter: &wave28P2PLimiter{autoOK: true},
			wallet:  wave28P2PWallet(t),
			wantErr: "parse params JSON",
		},
		{
			name: "price query error is wrapped",
			params: map[string]interface{}{
				"peer_did":  validDID.ID,
				"tool_name": "search_knowledge",
			},
			quoteErr:    errors.New("quote offline"),
			limiter:     &wave28P2PLimiter{autoOK: true},
			wallet:      wave28P2PWallet(t),
			wantErr:     "price query: quote offline",
			wantQueries: 1,
		},
		{
			name: "free invocation error is wrapped",
			params: map[string]interface{}{
				"peer_did":  validDID.ID,
				"tool_name": "search_knowledge",
			},
			quote:       &protocol.PriceQuoteResult{ToolName: "search_knowledge", IsFree: true},
			invokeErr:   errors.New("remote refused"),
			limiter:     &wave28P2PLimiter{autoOK: true},
			wallet:      wave28P2PWallet(t),
			wantErr:     "invoke free tool: remote refused",
			wantQueries: 1,
		},
		{
			name: "invalid paid quote price is wrapped",
			params: map[string]interface{}{
				"peer_did":  validDID.ID,
				"tool_name": "paid_tool",
			},
			quote:       wave28P2PPaidQuote("not-usdc"),
			limiter:     &wave28P2PLimiter{autoOK: true},
			wallet:      wave28P2PWallet(t),
			wantErr:     `parse price "not-usdc"`,
			wantQueries: 1,
		},
		{
			name: "spending limit error is wrapped",
			params: map[string]interface{}{
				"peer_did":  validDID.ID,
				"tool_name": "paid_tool",
			},
			quote:       wave28P2PPaidQuote("0.42"),
			limiter:     &wave28P2PLimiter{autoOK: true, checkErr: errors.New("daily cap")},
			wallet:      wave28P2PWallet(t),
			wantErr:     "spending limit: daily cap",
			wantQueries: 1,
		},
		{
			name: "auto approve error is wrapped",
			params: map[string]interface{}{
				"peer_did":  validDID.ID,
				"tool_name": "paid_tool",
			},
			quote:       wave28P2PPaidQuote("0.42"),
			limiter:     &wave28P2PLimiter{autoErr: errors.New("threshold unavailable")},
			wallet:      wave28P2PWallet(t),
			wantErr:     "auto-approve check: threshold unavailable",
			wantQueries: 1,
		},
		{
			name: "wallet address error is wrapped",
			params: map[string]interface{}{
				"peer_did":  validDID.ID,
				"tool_name": "paid_tool",
			},
			quote:       wave28P2PPaidQuote("0.42"),
			limiter:     &wave28P2PLimiter{autoOK: true},
			wallet:      wave28P2PWalletWithAddressError(t, errors.New("locked")),
			wantErr:     "get wallet address: locked",
			wantQueries: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			remote.reset()
			remote.quote = tt.quote
			remote.quoteErr = tt.quoteErr
			remote.invokeErr = tt.invokeErr

			tools := buildP2PPaidInvokeTool(&p2pComponents{sessions: sessions}, &paymentComponents{
				wallet:  tt.wallet,
				limiter: tt.limiter,
				chainID: 84532,
			})
			require.Len(t, tools, 1)

			got, err := tools[0].Handler(context.Background(), tt.params)

			require.Error(t, err)
			assert.Nil(t, got)
			assert.ErrorContains(t, err, tt.wantErr)
			assert.Equal(t, tt.wantQueries, remote.queryCalls)
			assert.False(t, remote.paidInvoked)
		})
	}
}

func TestWave28P2PPaidInvoke_SuccessRecordsSpendingAndMapsAuthorization(t *testing.T) {
	originalRemoteAgent := newP2PPaidInvokeRemoteAgent
	originalNewUnsigned := newEIP3009Unsigned
	t.Cleanup(func() {
		newP2PPaidInvokeRemoteAgent = originalRemoteAgent
		newEIP3009Unsigned = originalNewUnsigned
	})

	validDID := wave28P2PPeerDID(t)
	sessions := wave28P2PSessions(t, validDID.ID)
	remote := &wave28P2PRemoteAgent{
		quote: wave28P2PPaidQuote("1.25"),
		paidResponse: &protocol.Response{
			Status: protocol.ResponseStatusOK,
			Result: map[string]interface{}{"answer": "paid-result"},
		},
	}
	newP2PPaidInvokeRemoteAgent = func(string, *identity.DID, *handshake.Session, *p2pComponents) p2pPaidInvokeRemoteAgent {
		return remote
	}
	newEIP3009Unsigned = func(from, to common.Address, value *big.Int, _ time.Time) (*eip3009.UnsignedAuth, error) {
		return &eip3009.UnsignedAuth{
			From:        from,
			To:          to,
			Value:       new(big.Int).Set(value),
			ValidAfter:  big.NewInt(100),
			ValidBefore: big.NewInt(200),
			Nonce:       [32]byte{31: 7},
		}, nil
	}

	limiter := &wave28P2PLimiter{autoOK: true}
	wallet := wave28P2PWallet(t)
	tools := buildP2PPaidInvokeTool(&p2pComponents{sessions: sessions}, &paymentComponents{
		wallet:  wallet,
		limiter: limiter,
		chainID: 84532,
	})
	require.Len(t, tools, 1)

	got, err := tools[0].Handler(context.Background(), map[string]interface{}{
		"peer_did":  validDID.ID,
		"tool_name": "paid_tool",
		"params":    `{"query":"wave28"}`,
	})

	require.NoError(t, err)
	payload := wave28P2PPayload(t, got)
	assert.Equal(t, "ok", payload["status"])
	assert.Equal(t, true, payload["paid"])
	assert.Equal(t, "1.25", payload["price"])
	assert.Equal(t, "USDC", payload["currency"])
	assert.Equal(t, map[string]interface{}{"answer": "paid-result"}, payload["result"])

	require.True(t, remote.paidInvoked)
	assert.Equal(t, map[string]interface{}{"query": "wave28"}, remote.paidParams)
	assert.Equal(t, "1250000", remote.lastAuth["value"])
	assert.Equal(t, wallet.address, remote.lastAuth["from"])
	assert.Equal(t, common.HexToAddress(wave28P2PSellerAddress).Hex(), remote.lastAuth["to"])
	assert.Equal(t, "100", remote.lastAuth["validAfter"])
	assert.Equal(t, "200", remote.lastAuth["validBefore"])
	assert.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000000007", remote.lastAuth["nonce"])
	assert.Contains(t, []float64{27, 28}, remote.lastAuth["v"])
	assert.Len(t, strings.TrimPrefix(remote.lastAuth["r"].(string), "0x"), 64)
	assert.Len(t, strings.TrimPrefix(remote.lastAuth["s"].(string), "0x"), 64)
	require.Len(t, limiter.recorded, 1)
	assert.Equal(t, "1250000", limiter.recorded[0].String())
	assert.Equal(t, 1, wallet.signTransactionCalls)
}

const (
	wave28P2PPeerPrivateKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	wave28P2PWalletKey      = "1111111111111111111111111111111111111111111111111111111111111111"
	wave28P2PSellerAddress  = "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

type wave28P2PRemoteAgent struct {
	quote        *protocol.PriceQuoteResult
	quoteErr     error
	invokeErr    error
	invokeResult map[string]interface{}
	paidResponse *protocol.Response
	queryCalls   int
	paidInvoked  bool
	paidParams   map[string]interface{}
	lastAuth     map[string]interface{}
}

func (a *wave28P2PRemoteAgent) reset() {
	a.quote = nil
	a.quoteErr = nil
	a.invokeErr = nil
	a.invokeResult = nil
	a.paidResponse = nil
	a.queryCalls = 0
	a.paidInvoked = false
	a.paidParams = nil
	a.lastAuth = nil
}

func (a *wave28P2PRemoteAgent) QueryPrice(context.Context, string) (*protocol.PriceQuoteResult, error) {
	a.queryCalls++
	if a.quoteErr != nil {
		return nil, a.quoteErr
	}
	return a.quote, nil
}

func (a *wave28P2PRemoteAgent) InvokeTool(context.Context, string, map[string]interface{}) (map[string]interface{}, error) {
	if a.invokeErr != nil {
		return nil, a.invokeErr
	}
	return a.invokeResult, nil
}

func (a *wave28P2PRemoteAgent) InvokeToolPaid(_ context.Context, _ string, params map[string]interface{}, auth map[string]interface{}) (*protocol.Response, error) {
	a.paidInvoked = true
	a.paidParams = params
	a.lastAuth = auth
	if a.invokeErr != nil {
		return nil, a.invokeErr
	}
	return a.paidResponse, nil
}

type wave28P2PLimiter struct {
	checkErr error
	autoErr  error
	autoOK   bool
	recorded []*big.Int
}

func (l *wave28P2PLimiter) Check(context.Context, *big.Int) error {
	return l.checkErr
}

func (l *wave28P2PLimiter) Record(_ context.Context, amount *big.Int) error {
	l.recorded = append(l.recorded, new(big.Int).Set(amount))
	return nil
}

func (*wave28P2PLimiter) DailySpent(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (*wave28P2PLimiter) DailyRemaining(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (l *wave28P2PLimiter) IsAutoApprovable(context.Context, *big.Int) (bool, error) {
	if l.autoErr != nil {
		return false, l.autoErr
	}
	return l.autoOK, nil
}

type wave28P2PTestWallet struct {
	key                  *ecdsa.PrivateKey
	address              string
	addressErr           error
	signTransactionCalls int
}

func (w *wave28P2PTestWallet) Address(context.Context) (string, error) {
	if w.addressErr != nil {
		return "", w.addressErr
	}
	return w.address, nil
}

func (*wave28P2PTestWallet) Balance(context.Context) (*big.Int, error) {
	return big.NewInt(0), nil
}

func (w *wave28P2PTestWallet) SignTransaction(_ context.Context, rawTx []byte) ([]byte, error) {
	w.signTransactionCalls++
	return ethcrypto.Sign(rawTx, w.key)
}

func (*wave28P2PTestWallet) SignMessage(context.Context, []byte) ([]byte, error) {
	return nil, nil
}

func (w *wave28P2PTestWallet) PublicKey(context.Context) ([]byte, error) {
	return ethcrypto.CompressPubkey(&w.key.PublicKey), nil
}

func wave28P2PWallet(t *testing.T) *wave28P2PTestWallet {
	t.Helper()

	key, err := ethcrypto.HexToECDSA(wave28P2PWalletKey)
	require.NoError(t, err)
	return &wave28P2PTestWallet{
		key:     key,
		address: ethcrypto.PubkeyToAddress(key.PublicKey).Hex(),
	}
}

func wave28P2PWalletWithAddressError(t *testing.T, err error) *wave28P2PTestWallet {
	t.Helper()

	wallet := wave28P2PWallet(t)
	wallet.addressErr = err
	return wallet
}

func wave28P2PPeerDID(t *testing.T) *identity.DID {
	t.Helper()

	key, err := ethcrypto.HexToECDSA(wave28P2PPeerPrivateKey)
	require.NoError(t, err)
	did, err := identity.DIDFromPublicKey(ethcrypto.CompressPubkey(&key.PublicKey))
	require.NoError(t, err)
	return did
}

func wave28P2PSessions(t *testing.T, peerDIDs ...string) *handshake.SessionStore {
	t.Helper()

	sessions, err := handshake.NewSessionStore(time.Hour)
	require.NoError(t, err)
	for _, peerDID := range peerDIDs {
		_, err := sessions.Create(peerDID, true)
		require.NoError(t, err)
	}
	return sessions
}

func wave28P2PPaidQuote(price string) *protocol.PriceQuoteResult {
	return &protocol.PriceQuoteResult{
		ToolName:   "paid_tool",
		Price:      price,
		Currency:   "USDC",
		SellerAddr: wave28P2PSellerAddress,
		ChainID:    84532,
		IsFree:     false,
	}
}

func wave28P2PRequiredParams(t *testing.T, tool *agent.Tool) []string {
	t.Helper()

	required, ok := tool.Parameters["required"].([]string)
	require.True(t, ok)
	return required
}

func wave28P2PParamType(t *testing.T, tool *agent.Tool, name string) string {
	t.Helper()

	properties, ok := tool.Parameters["properties"].(map[string]interface{})
	require.True(t, ok)
	property, ok := properties[name].(map[string]interface{})
	require.True(t, ok)
	paramType, ok := property["type"].(string)
	require.True(t, ok)
	return paramType
}

func wave28P2PPayload(t *testing.T, result interface{}) map[string]interface{} {
	t.Helper()

	payload, ok := result.(map[string]interface{})
	require.True(t, ok)
	return payload
}
