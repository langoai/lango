package app

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/langoai/lango/internal/gateway"
	"github.com/langoai/lango/internal/wallet"
)

// writeJSON encodes v as JSON into the response writer.
func writeJSON(w http.ResponseWriter, v interface{}) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// registerP2PRoutes mounts P2P status endpoints on the gateway router.
// When OIDC is configured, endpoints require authentication.
// When auth is nil (dev mode), endpoints are accessible without authentication.
func registerP2PRoutes(r chi.Router, p2pc *p2pComponents, auth *gateway.AuthManager) {
	r.Route("/api/p2p", func(r chi.Router) {
		r.Use(gateway.RequireAuth(auth))
		r.Get("/status", p2pStatusHandler(p2pc))
		r.Get("/peers", p2pPeersHandler(p2pc))
		r.Get("/identity", p2pIdentityHandler(p2pc))
		r.Get("/reputation", p2pReputationHandler(p2pc))
		r.Get("/pricing", p2pPricingHandler(p2pc))
	})
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

		if p2pc.identity == nil {
			writeJSON(w, map[string]interface{}{
				"did":    nil,
				"peerId": p2pc.node.PeerID().String(),
			})
			return
		}

		ctx := context.Background()
		did, err := p2pc.identity.DID(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			writeJSON(w, map[string]string{
				"error": err.Error(),
			})
			return
		}

		writeJSON(w, map[string]interface{}{
			"did":    did.ID,
			"peerId": p2pc.node.PeerID().String(),
		})
	}
}
