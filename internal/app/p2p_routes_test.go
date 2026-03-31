package app

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/go-chi/chi/v5"
	"github.com/langoai/lango/internal/config"
	p2pnet "github.com/langoai/lango/internal/p2p"
	"github.com/langoai/lango/internal/p2p/handshake"
	"github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/p2p/provenanceproto"
	provenancepkg "github.com/langoai/lango/internal/provenance"
	"github.com/libp2p/go-libp2p"
	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- p2pPricingHandler ---

func TestP2PPricingHandler_AllPrices(t *testing.T) {
	p2pc := &p2pComponents{
		pricingCfg: config.P2PPricingConfig{
			Enabled:    true,
			PerQuery:   "0.50",
			ToolPrices: map[string]string{"web_search": "1.00", "code_exec": "2.00"},
		},
	}

	handler := p2pPricingHandler(p2pc)
	req := httptest.NewRequest("GET", "/api/p2p/pricing", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, true, resp["enabled"])
	assert.Equal(t, "0.50", resp["perQuery"])
	assert.Equal(t, "USDC", resp["currency"])

	toolPrices, ok := resp["toolPrices"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "1.00", toolPrices["web_search"])
	assert.Equal(t, "2.00", toolPrices["code_exec"])
}

func TestP2PPricingHandler_SpecificTool(t *testing.T) {
	p2pc := &p2pComponents{
		pricingCfg: config.P2PPricingConfig{
			PerQuery:   "0.50",
			ToolPrices: map[string]string{"web_search": "1.00"},
		},
	}

	handler := p2pPricingHandler(p2pc)
	req := httptest.NewRequest("GET", "/api/p2p/pricing?tool=web_search", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "web_search", resp["tool"])
	assert.Equal(t, "1.00", resp["price"])
	assert.Equal(t, "USDC", resp["currency"])
}

func TestP2PPricingHandler_UnknownToolFallsBackToPerQuery(t *testing.T) {
	p2pc := &p2pComponents{
		pricingCfg: config.P2PPricingConfig{
			PerQuery:   "0.50",
			ToolPrices: map[string]string{"web_search": "1.00"},
		},
	}

	handler := p2pPricingHandler(p2pc)
	req := httptest.NewRequest("GET", "/api/p2p/pricing?tool=unknown_tool", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "unknown_tool", resp["tool"])
	assert.Equal(t, "0.50", resp["price"], "should fall back to perQuery price")
}

func TestP2PPricingHandler_Disabled(t *testing.T) {
	p2pc := &p2pComponents{
		pricingCfg: config.P2PPricingConfig{
			Enabled:  false,
			PerQuery: "0.00",
		},
	}

	handler := p2pPricingHandler(p2pc)
	req := httptest.NewRequest("GET", "/api/p2p/pricing", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, false, resp["enabled"])
}

// --- p2pReputationHandler ---

func TestP2PReputationHandler_MissingPeerDID(t *testing.T) {
	p2pc := &p2pComponents{}

	handler := p2pReputationHandler(p2pc)
	req := httptest.NewRequest("GET", "/api/p2p/reputation", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]string
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "peer_did")
}

func TestP2PReputationHandler_NilReputationSystem(t *testing.T) {
	p2pc := &p2pComponents{
		reputation: nil,
	}

	handler := p2pReputationHandler(p2pc)
	req := httptest.NewRequest("GET", "/api/p2p/reputation?peer_did=did:lango:abc123", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp map[string]string
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Contains(t, resp["error"], "not available")
}

// --- p2pIdentityHandler ---

func TestP2PIdentityHandler_NilIdentity(t *testing.T) {
	// When identity is nil but node is also nil, handler will panic at node.PeerID().
	// We test only the nil identity path by providing a minimal node.
	// Since creating a real node requires libp2p, this test documents the expected behavior.
	t.Skip("requires libp2p node; tested via integration tests")
}

type routeTestWallet struct {
	key *ecdsa.PrivateKey
}

func newRouteTestWallet(t *testing.T) (*routeTestWallet, string) {
	t.Helper()
	key, err := ethcrypto.GenerateKey()
	require.NoError(t, err)
	pub := ethcrypto.CompressPubkey(&key.PublicKey)
	did, err := identity.DIDFromPublicKey(pub)
	require.NoError(t, err)
	return &routeTestWallet{key: key}, did.ID
}

func (w *routeTestWallet) Address(context.Context) (string, error)   { return "0x0", nil }
func (w *routeTestWallet) Balance(context.Context) (*big.Int, error) { return big.NewInt(0), nil }
func (w *routeTestWallet) SignTransaction(context.Context, []byte) ([]byte, error) {
	return nil, nil
}
func (w *routeTestWallet) SignMessage(_ context.Context, message []byte) ([]byte, error) {
	return ethcrypto.Sign(ethcrypto.Keccak256(message), w.key)
}
func (w *routeTestWallet) PublicKey(context.Context) ([]byte, error) {
	return ethcrypto.CompressPubkey(&w.key.PublicKey), nil
}

func setupProvenanceRouteRuntime(t *testing.T) (*App, *p2pComponents, host.Host, string, func()) {
	t.Helper()

	cfg := config.DefaultConfig()
	cfg.P2P.Enabled = true
	cfg.P2P.KeyDir = t.TempDir() //nolint:staticcheck // testing deprecated field for backward compat
	cfg.P2P.ListenAddrs = []string{"/ip4/127.0.0.1/tcp/0"}
	cfg.P2P.MaxPeers = 8

	node, err := p2pnet.NewNode(cfg.P2P, testLog(), nil)
	require.NoError(t, err)

	sessions, err := handshake.NewSessionStore(time.Hour)
	require.NoError(t, err)

	walletProvider, localDID := newRouteTestWallet(t)
	idProvider := identity.NewProvider(walletProvider, testLog())

	cpStore := provenancepkg.NewMemoryStore()
	treeStore := provenancepkg.NewMemoryTreeStore()
	attrStore := provenancepkg.NewMemoryAttributionStore()
	attrSvc := provenancepkg.NewAttributionService(attrStore, cpStore, nil)
	bundleSvc := provenancepkg.NewBundleService(cpStore, treeStore, attrStore, attrSvc)
	require.NoError(t, cpStore.SaveCheckpoint(context.Background(), provenancepkg.Checkpoint{
		ID:         "cp-1",
		SessionKey: "sess-1",
		Label:      "checkpoint",
		Trigger:    provenancepkg.TriggerManual,
		CreatedAt:  time.Now(),
	}))

	app := &App{
		Config:                cfg,
		WalletProvider:        walletProvider,
		ProvenanceBundle:      bundleSvc,
		ProvenanceSessionTree: provenancepkg.NewSessionTree(treeStore),
	}

	p2pc := &p2pComponents{
		node:     node,
		sessions: sessions,
		identity: idProvider,
	}

	serverKey, _, err := libp2pcrypto.GenerateSecp256k1Key(nil)
	require.NoError(t, err)
	serverHost, err := libp2p.New(libp2p.Identity(serverKey))
	require.NoError(t, err)

	pub, err := serverKey.GetPublic().Raw()
	require.NoError(t, err)
	serverDID, err := identity.DIDFromPublicKey(pub)
	require.NoError(t, err)

	require.NoError(t, node.Host().Connect(context.Background(), peer.AddrInfo{
		ID:    serverHost.ID(),
		Addrs: serverHost.Addrs(),
	}))
	_, err = sessions.Create(serverDID.ID, true)
	require.NoError(t, err)

	cleanup := func() {
		_ = node.Stop()
		_ = serverHost.Close()
		_ = localDID
	}

	return app, p2pc, serverHost, serverDID.ID, cleanup
}

func TestP2PProvenancePushHandler_InvalidRedaction(t *testing.T) {
	app, p2pc, _, _, cleanup := setupProvenanceRouteRuntime(t)
	defer cleanup()

	router := chi.NewRouter()
	registerP2PRoutes(router, app, p2pc, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/p2p/provenance/push",
		strings.NewReader(`{"peerDid":"did:lango:abc","sessionKey":"sess-1","redaction":"bogus"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid redaction level")
}

func TestP2PProvenanceFetchHandler_InvalidRedaction(t *testing.T) {
	app, p2pc, _, _, cleanup := setupProvenanceRouteRuntime(t)
	defer cleanup()

	router := chi.NewRouter()
	registerP2PRoutes(router, app, p2pc, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/p2p/provenance/fetch",
		strings.NewReader(`{"peerDid":"did:lango:abc","sessionKey":"sess-1","redaction":"invalid"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid redaction level")
}

func TestP2PProvenancePushHandler_RequiresActiveSession(t *testing.T) {
	app, p2pc, _, _, cleanup := setupProvenanceRouteRuntime(t)
	defer cleanup()

	router := chi.NewRouter()
	registerP2PRoutes(router, app, p2pc, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/p2p/provenance/push",
		strings.NewReader(`{"peerDid":"did:lango:missing","sessionKey":"sess-1","redaction":"content"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "active session required")
}

func TestP2PProvenancePushAndFetchHandlers(t *testing.T) {
	app, p2pc, serverHost, serverPeerDID, cleanup := setupProvenanceRouteRuntime(t)
	defer cleanup()

	remoteWallet, remoteSignerDID := newRouteTestWallet(t)
	remoteCP := provenancepkg.NewMemoryStore()
	remoteTree := provenancepkg.NewMemoryTreeStore()
	remoteAttrs := provenancepkg.NewMemoryAttributionStore()
	remoteAttrSvc := provenancepkg.NewAttributionService(remoteAttrs, remoteCP, nil)
	remoteBundleSvc := provenancepkg.NewBundleService(remoteCP, remoteTree, remoteAttrs, remoteAttrSvc)
	require.NoError(t, remoteCP.SaveCheckpoint(context.Background(), provenancepkg.Checkpoint{
		ID:         "remote-cp",
		SessionKey: "sess-remote",
		Label:      "remote",
		Trigger:    provenancepkg.TriggerManual,
		CreatedAt:  time.Now(),
	}))

	var pushed []byte
	handler := provenanceproto.NewHandler(provenanceproto.HandlerConfig{
		Validator: func(token string) (string, bool) {
			return serverPeerDID, token != ""
		},
		Importer: func(_ context.Context, peerDID string, data []byte) error {
			assert.Equal(t, serverPeerDID, peerDID)
			pushed = append([]byte(nil), data...)
			return nil
		},
		Exporter: func(ctx context.Context, peerDID, sessionKey, redaction string) ([]byte, error) {
			assert.Equal(t, serverPeerDID, peerDID)
			_, data, err := remoteBundleSvc.Export(ctx, sessionKey, provenancepkg.RedactionLevel(redaction), remoteSignerDID, func(ctx context.Context, payload []byte) ([]byte, error) {
				return remoteWallet.SignMessage(ctx, payload)
			})
			return data, err
		},
	})
	serverHost.SetStreamHandler(provenanceproto.ProtocolID, handler.StreamHandler())

	router := chi.NewRouter()
	registerP2PRoutes(router, app, p2pc, nil)

	pushReq := httptest.NewRequest(http.MethodPost, "/api/p2p/provenance/push",
		strings.NewReader(fmt.Sprintf(`{"peerDid":"%s","sessionKey":"sess-1","redaction":"content"}`, serverPeerDID)))
	pushReq.Header.Set("Content-Type", "application/json")
	pushW := httptest.NewRecorder()
	router.ServeHTTP(pushW, pushReq)
	require.Equal(t, http.StatusOK, pushW.Code, pushW.Body.String())
	require.NotEmpty(t, pushed)

	fetchReq := httptest.NewRequest(http.MethodPost, "/api/p2p/provenance/fetch",
		strings.NewReader(fmt.Sprintf(`{"peerDid":"%s","sessionKey":"sess-remote","redaction":"content"}`, serverPeerDID)))
	fetchReq.Header.Set("Content-Type", "application/json")
	fetchW := httptest.NewRecorder()
	router.ServeHTTP(fetchW, fetchReq)
	require.Equal(t, http.StatusOK, fetchW.Code, fetchW.Body.String())
	assert.Contains(t, fetchW.Body.String(), remoteSignerDID)
}

func TestP2PProvenanceFetchHandler_TamperedBundle(t *testing.T) {
	app, p2pc, serverHost, serverPeerDID, cleanup := setupProvenanceRouteRuntime(t)
	defer cleanup()

	handler := provenanceproto.NewHandler(provenanceproto.HandlerConfig{
		Validator: func(token string) (string, bool) {
			return serverPeerDID, token != ""
		},
		Exporter: func(_ context.Context, peerDID, sessionKey, redaction string) ([]byte, error) {
			return []byte(`{"version":"1","signer_did":"did:lango:deadbeef","signature_algorithm":"secp256k1-keccak256","signature":"AQ=="}`), nil
		},
	})
	serverHost.SetStreamHandler(provenanceproto.ProtocolID, handler.StreamHandler())

	router := chi.NewRouter()
	registerP2PRoutes(router, app, p2pc, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/p2p/provenance/fetch",
		strings.NewReader(fmt.Sprintf(`{"peerDid":"%s","sessionKey":"sess-remote","redaction":"content"}`, serverPeerDID)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "import provenance bundle")
}
