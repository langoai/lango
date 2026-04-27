package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/langoai/lango/internal/gateway"
	"github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/p2p/provenanceproto"
	"github.com/langoai/lango/internal/provenance"
	"github.com/langoai/lango/internal/wallet"
)

// writeJSON encodes v as JSON into the response writer.
func writeJSON(w http.ResponseWriter, v interface{}) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// registerP2PRoutes mounts P2P status endpoints on the gateway router.
// The subtree is public only when gateway auth is disabled; otherwise the
// RequireAuth middleware protects every /api/p2p route.
func registerP2PRoutes(r chi.Router, app *App, p2pc *p2pComponents, auth *gateway.AuthManager) {
	r.Route("/api/p2p", func(r chi.Router) {
		r.Use(gateway.RequireAuth(auth))
		r.Get("/status", p2pStatusHandler(p2pc))
		r.Get("/peers", p2pPeersHandler(p2pc))
		r.Get("/identity", p2pIdentityHandler(p2pc))
		r.Get("/reputation", p2pReputationHandler(p2pc))
		r.Get("/pricing", p2pPricingHandler(p2pc))
		r.Post("/provenance/push", p2pProvenancePushHandler(app, p2pc))
		r.Post("/provenance/fetch", p2pProvenanceFetchHandler(app, p2pc))
	})
}

type provenanceExchangeRequest struct {
	PeerDID    string `json:"peerDid"`
	SessionKey string `json:"sessionKey"`
	Redaction  string `json:"redaction"`
}

func p2pStatusHandler(p2pc *p2pComponents) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		node := p2pc.node

		addrs := make([]string, 0, len(node.Multiaddrs()))
		for _, a := range node.Multiaddrs() {
			addrs = append(addrs, a.String())
		}

		resp := map[string]interface{}{
			"peerId":         node.PeerID().String(),
			"listenAddrs":    addrs,
			"connectedPeers": len(node.ConnectedPeers()),
			"mdnsEnabled":    p2pc.node.Host().Addrs() != nil,
		}

		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, resp)
	}
}

func p2pPeersHandler(p2pc *p2pComponents) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		node := p2pc.node
		connected := node.ConnectedPeers()

		type peerInfo struct {
			PeerID string   `json:"peerId"`
			Addrs  []string `json:"addrs"`
		}

		peers := make([]peerInfo, 0, len(connected))
		for _, pid := range connected {
			conns := node.Host().Network().ConnsToPeer(pid)
			var addrs []string
			for _, c := range conns {
				addrs = append(addrs, c.RemoteMultiaddr().String())
			}
			peers = append(peers, peerInfo{
				PeerID: pid.String(),
				Addrs:  addrs,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		writeJSON(w, map[string]interface{}{
			"peers": peers,
			"count": len(peers),
		})
	}
}

func p2pReputationHandler(p2pc *p2pComponents) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		peerDID := r.URL.Query().Get("peer_did")
		if peerDID == "" {
			w.WriteHeader(http.StatusBadRequest)
			writeJSON(w, map[string]string{
				"error": "peer_did query parameter is required",
			})
			return
		}

		if p2pc.reputation == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			writeJSON(w, map[string]string{
				"error": "reputation system not available",
			})
			return
		}

		details, err := p2pc.reputation.GetDetails(r.Context(), peerDID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeJSON(w, map[string]string{
				"error": err.Error(),
			})
			return
		}

		if details == nil {
			writeJSON(w, map[string]interface{}{
				"peerDid":    peerDID,
				"trustScore": 0.0,
				"message":    "no reputation record found",
			})
			return
		}

		writeJSON(w, details)
	}
}

func p2pPricingHandler(p2pc *p2pComponents) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		pricing := p2pc.pricingCfg
		toolName := r.URL.Query().Get("tool")

		if toolName != "" {
			price, ok := pricing.ToolPrices[toolName]
			if !ok {
				price = pricing.PerQuery
			}
			writeJSON(w, map[string]interface{}{
				"tool":     toolName,
				"price":    price,
				"currency": wallet.CurrencyUSDC,
			})
			return
		}

		writeJSON(w, map[string]interface{}{
			"enabled":    pricing.Enabled,
			"perQuery":   pricing.PerQuery,
			"toolPrices": pricing.ToolPrices,
			"currency":   wallet.CurrencyUSDC,
		})
	}
}

func p2pIdentityHandler(p2pc *p2pComponents) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var did any
		if p2pc.identity == nil {
			writeJSON(w, map[string]any{
				"did":    did,
				"peerId": p2pc.node.PeerID().String(),
			})
			return
		}

		ctx := r.Context()
		if resolvedDID, err := p2pc.identity.DID(ctx); err == nil && resolvedDID != nil && resolvedDID.ID != "" {
			did = resolvedDID.ID
		}

		writeJSON(w, map[string]any{
			"did":    did,
			"peerId": p2pc.node.PeerID().String(),
		})
	}
}

func p2pProvenancePushHandler(app *App, p2pc *p2pComponents) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		req, ok := decodeProvenanceRequest(w, r)
		if !ok {
			return
		}
		token, target, ok := resolveProvenancePeer(w, req.PeerDID, p2pc)
		if !ok {
			return
		}
		did, signFn, ok := provenanceSigner(w, r.Context(), app, p2pc)
		if !ok {
			return
		}

		_, bundle, err := app.ProvenanceBundle.Export(r.Context(), req.SessionKey, provenance.RedactionLevel(req.Redaction), did, signFn)
		if err != nil {
			http.Error(w, "export provenance bundle: "+err.Error(), http.StatusBadRequest)
			return
		}

		resp, err := provenanceproto.PushBundle(r.Context(), p2pc.node.Host(), target.PeerID, token, bundle)
		if err != nil {
			http.Error(w, "push provenance bundle: "+err.Error(), http.StatusBadGateway)
			return
		}
		writeJSON(w, map[string]any{
			"pushed":    resp.Stored,
			"peerDid":   req.PeerDID,
			"message":   resp.Message,
			"redaction": req.Redaction,
		})
	}
}

func p2pProvenanceFetchHandler(app *App, p2pc *p2pComponents) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		req, ok := decodeProvenanceRequest(w, r)
		if !ok {
			return
		}
		token, target, ok := resolveProvenancePeer(w, req.PeerDID, p2pc)
		if !ok {
			return
		}

		data, err := provenanceproto.FetchBundle(r.Context(), p2pc.node.Host(), target.PeerID, token, req.SessionKey, req.Redaction)
		if err != nil {
			http.Error(w, "fetch provenance bundle: "+err.Error(), http.StatusBadGateway)
			return
		}
		bundle, err := app.ProvenanceBundle.Import(r.Context(), data)
		if err != nil {
			http.Error(w, "import provenance bundle: "+err.Error(), http.StatusBadRequest)
			return
		}

		writeJSON(w, map[string]any{
			"imported":  true,
			"peerDid":   req.PeerDID,
			"signerDid": bundle.SignerDID,
			"redaction": bundle.RedactionLevel,
		})
	}
}

func decodeProvenanceRequest(w http.ResponseWriter, r *http.Request) (*provenanceExchangeRequest, bool) {
	var req provenanceExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "decode request: "+err.Error(), http.StatusBadRequest)
		return nil, false
	}
	if req.PeerDID == "" || req.SessionKey == "" {
		http.Error(w, "peerDid and sessionKey are required", http.StatusBadRequest)
		return nil, false
	}
	if req.Redaction == "" {
		req.Redaction = string(provenance.RedactionContent)
	}
	if !provenance.RedactionLevel(req.Redaction).Valid() {
		http.Error(w, fmt.Sprintf("invalid redaction level %q: must be none, content, or full", req.Redaction), http.StatusBadRequest)
		return nil, false
	}
	return &req, true
}

func resolveProvenancePeer(w http.ResponseWriter, peerDID string, p2pc *p2pComponents) (string, *identity.DID, bool) {
	if p2pc == nil || p2pc.sessions == nil || p2pc.node == nil {
		http.Error(w, "P2P runtime is not available", http.StatusServiceUnavailable)
		return "", nil, false
	}
	sess := p2pc.sessions.Get(peerDID)
	if sess == nil {
		http.Error(w, "active session required for peer DID", http.StatusConflict)
		return "", nil, false
	}
	target, err := identity.ParseDID(peerDID)
	if err != nil {
		http.Error(w, "parse peer DID: "+err.Error(), http.StatusBadRequest)
		return "", nil, false
	}
	return sess.Token, target, true
}

func provenanceSigner(w http.ResponseWriter, ctx context.Context, app *App, p2pc *p2pComponents) (string, provenance.BundleSigner, bool) {
	if app == nil || app.ProvenanceBundle == nil || app.WalletProvider == nil || p2pc == nil || p2pc.identity == nil {
		http.Error(w, "local signed provenance export requires wallet identity and provenance bundle service", http.StatusServiceUnavailable)
		return "", nil, false
	}
	// Use the wallet's v1 DID for provenance signing. The wallet signer uses
	// secp256k1-keccak256, and VerifyMessageSignature only supports v1 DIDs.
	// Using p2pc.identity.DID() would return a v2 DID when BundleProvider is
	// active, causing verification failures on the receiving end.
	walletPub, err := app.WalletProvider.PublicKey(ctx)
	if err != nil {
		http.Error(w, "resolve wallet public key: "+err.Error(), http.StatusServiceUnavailable)
		return "", nil, false
	}
	walletDID, err := identity.DIDFromPublicKey(walletPub)
	if err != nil {
		http.Error(w, "derive wallet DID: "+err.Error(), http.StatusServiceUnavailable)
		return "", nil, false
	}
	return walletDID.ID, &walletBundleSigner{wp: app.WalletProvider}, true
}
