package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/langoai/lango/internal/config"
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
