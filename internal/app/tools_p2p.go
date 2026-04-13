package app

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/p2p/discovery"
	"github.com/langoai/lango/internal/toolparam"
	"github.com/langoai/lango/internal/p2p/firewall"
	"github.com/langoai/lango/internal/p2p/handshake"
	"github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/p2p/protocol"
	"github.com/langoai/lango/internal/payment"
	"github.com/langoai/lango/internal/payment/contracts"
	"github.com/langoai/lango/internal/payment/eip3009"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/wallet"
	"github.com/libp2p/go-libp2p/core/peer"
	libp2pproto "github.com/libp2p/go-libp2p/core/protocol"
	"github.com/multiformats/go-multiaddr"
)

// buildP2PTools creates P2P networking tools.
func buildP2PTools(pc *p2pComponents) []*agent.Tool {
	return []*agent.Tool{
		{
			Name:        "p2p_status",
			Description: "Show P2P node status: peer ID, listen addresses, connected peers",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "p2p",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				addrs := pc.node.Multiaddrs()
				addrStrs := make([]string, len(addrs))
				for i, a := range addrs {
					addrStrs[i] = a.String()
				}
				connected := pc.node.ConnectedPeers()
				peerStrs := make([]string, len(connected))
				for i, p := range connected {
					peerStrs[i] = p.String()
				}

				// Get local DID if available.
				var did string
				if pc.identity != nil {
					d, err := pc.identity.DID(ctx)
					if err == nil && d != nil {
						did = d.ID
					}
				}

				return map[string]interface{}{
					"peerID":         pc.node.PeerID().String(),
					"did":            did,
					"listenAddrs":    addrStrs,
					"connectedPeers": peerStrs,
					"peerCount":      len(connected),
					"sessions":       len(pc.sessions.ActiveSessions()),
				}, nil
			},
		},
		{
			Name:        "p2p_connect",
			Description: "Initiate a handshake with a remote peer by multiaddr",
			SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Category: "p2p",
				Activity: agent.ActivityExecute,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"multiaddr": map[string]interface{}{"type": "string", "description": "The peer's multiaddr (e.g., /ip4/1.2.3.4/tcp/9000/p2p/QmPeer...)"},
				},
				"required": []string{"multiaddr"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				addr, err := toolparam.RequireString(params, "multiaddr")
				if err != nil {
					return nil, err
				}

				// Parse multiaddr and extract peer info.
				ma, err := multiaddr.NewMultiaddr(addr)
				if err != nil {
					return nil, fmt.Errorf("invalid multiaddr: %w", err)
				}
				pi, err := peer.AddrInfoFromP2pAddr(ma)
				if err != nil {
					return nil, fmt.Errorf("parse peer addr: %w", err)
				}

				// Connect to the peer.
				if err := pc.node.Host().Connect(ctx, *pi); err != nil {
					return nil, fmt.Errorf("connect to peer: %w", err)
				}

				// Open a handshake stream with protocol preference order.
				protocols := handshake.PreferredProtocols(pc.kemEnabled)
				protoIDs := make([]libp2pproto.ID, len(protocols))
				for i, p := range protocols {
					protoIDs[i] = libp2pproto.ID(p)
				}
				s, err := pc.node.Host().NewStream(ctx, pi.ID, protoIDs...)
				if err != nil {
					return nil, fmt.Errorf("open handshake stream: %w", err)
				}
				defer s.Close()

				// Get local DID.
				localDID := ""
				if pc.identity != nil {
					d, err := pc.identity.DID(ctx)
					if err == nil && d != nil {
						localDID = d.ID
					}
				}

				sess, err := pc.handshaker.Initiate(ctx, s, localDID)
				if err != nil {
					return nil, fmt.Errorf("handshake: %w", err)
				}

				return map[string]interface{}{
					"status":     "connected",
					"peerID":     pi.ID.String(),
					"peerDID":    sess.PeerDID,
					"zkVerified": sess.ZKVerified,
					"expiresAt":  sess.ExpiresAt.Format(time.RFC3339),
				}, nil
			},
		},
		{
			Name:        "p2p_disconnect",
			Description: "Disconnect from a peer",
			SafetyLevel: agent.SafetyLevelModerate,
			Capability: agent.ToolCapability{
				Category: "p2p",
				Activity: agent.ActivityManage,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"peer_did": map[string]interface{}{"type": "string", "description": "The peer's DID to disconnect"},
				},
				"required": []string{"peer_did"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				peerDID, err := toolparam.RequireString(params, "peer_did")
				if err != nil {
					return nil, err
				}
				pc.sessions.Remove(peerDID)
				return map[string]interface{}{
					"status":  "disconnected",
					"peerDID": peerDID,
				}, nil
			},
		},
		{
			Name:        "p2p_peers",
			Description: "List connected peers with session info",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "p2p",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				sessions := pc.sessions.ActiveSessions()
				peers := make([]map[string]interface{}, 0, len(sessions))
				for _, s := range sessions {
					peers = append(peers, map[string]interface{}{
						"peerDID":    s.PeerDID,
						"zkVerified": s.ZKVerified,
						"createdAt":  s.CreatedAt.Format(time.RFC3339),
						"expiresAt":  s.ExpiresAt.Format(time.RFC3339),
					})
				}
				return map[string]interface{}{"peers": peers, "count": len(peers)}, nil
			},
		},
		{
			Name:        "p2p_query",
			Description: "Send an inference-only query to a connected peer",
			SafetyLevel: agent.SafetyLevelModerate,
			Capability: agent.ToolCapability{
				Category: "p2p",
				Activity: agent.ActivityExecute,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"peer_did":  map[string]interface{}{"type": "string", "description": "The peer's DID to query"},
					"tool_name": map[string]interface{}{"type": "string", "description": "Tool to invoke on the remote agent"},
					"params":    map[string]interface{}{"type": "string", "description": "JSON string of parameters for the tool"},
				},
				"required": []string{"peer_did", "tool_name"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				peerDID, err := toolparam.RequireString(params, "peer_did")
				if err != nil {
					return nil, err
				}
				toolName, err := toolparam.RequireString(params, "tool_name")
				if err != nil {
					return nil, err
				}
				paramStr := toolparam.OptionalString(params, "params", "")

				sess := pc.sessions.Get(peerDID)
				if sess == nil {
					return nil, fmt.Errorf("no active session for peer %s", peerDID)
				}

				// Parse the peer ID from DID.
				did, err := identity.ParseDID(peerDID)
				if err != nil {
					return nil, fmt.Errorf("parse peer DID: %w", err)
				}

				var toolParams map[string]interface{}
				if paramStr != "" {
					if err := json.Unmarshal([]byte(paramStr), &toolParams); err != nil {
						return nil, fmt.Errorf("parse params JSON: %w", err)
					}
				}
				if toolParams == nil {
					toolParams = map[string]interface{}{}
				}

				remoteAgent := protocol.NewRemoteAgent(protocol.RemoteAgentConfig{
					Name:         "peer-" + peerDID[:16],
					DID:          peerDID,
					PeerID:       did.PeerID,
					SessionToken: sess.Token,
					Host:         pc.node.Host(),
					Logger:       logger(),
				})

				result, err := remoteAgent.InvokeTool(ctx, toolName, toolParams)
				if err != nil {
					return nil, fmt.Errorf("remote tool invoke: %w", err)
				}

				return result, nil
			},
		},
		{
			Name:        "p2p_firewall_rules",
			Description: "List current firewall ACL rules",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "p2p",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				rules := pc.fw.Rules()
				ruleList := make([]map[string]interface{}, len(rules))
				for i, r := range rules {
					ruleList[i] = map[string]interface{}{
						"peerDID":   r.PeerDID,
						"action":    r.Action,
						"tools":     r.Tools,
						"rateLimit": r.RateLimit,
					}
				}
				return map[string]interface{}{"rules": ruleList, "count": len(rules)}, nil
			},
		},
		{
			Name:        "p2p_firewall_add",
			Description: "Add a firewall ACL rule",
			SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Category: "p2p",
				Activity: agent.ActivityManage,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"peer_did":   map[string]interface{}{"type": "string", "description": "Peer DID to apply rule to (* for all)"},
					"action":     map[string]interface{}{"type": "string", "description": "allow or deny", "enum": []string{"allow", "deny"}},
					"tools":      map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Tool name patterns (* for all)"},
					"rate_limit": map[string]interface{}{"type": "integer", "description": "Max requests per minute (0 = unlimited)"},
				},
				"required": []string{"peer_did", "action"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				peerDID, err := toolparam.RequireString(params, "peer_did")
				if err != nil {
					return nil, err
				}
				action, err := toolparam.RequireString(params, "action")
				if err != nil {
					return nil, err
				}

				tools := toolparam.StringSlice(params, "tools")

				rateLimit := toolparam.OptionalInt(params, "rate_limit", 0)

				rule := firewall.ACLRule{
					PeerDID:   peerDID,
					Action:    firewall.ACLAction(action),
					Tools:     tools,
					RateLimit: rateLimit,
				}
				if err := pc.fw.AddRule(rule); err != nil {
					return nil, fmt.Errorf("add firewall rule: %w", err)
				}

				return map[string]interface{}{
					"status":  "added",
					"message": fmt.Sprintf("Firewall rule added: %s %s", action, peerDID),
				}, nil
			},
		},
		{
			Name:        "p2p_firewall_remove",
			Description: "Remove all firewall rules for a peer DID",
			SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Category: "p2p",
				Activity: agent.ActivityManage,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"peer_did": map[string]interface{}{"type": "string", "description": "Peer DID to remove rules for"},
				},
				"required": []string{"peer_did"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				peerDID, err := toolparam.RequireString(params, "peer_did")
				if err != nil {
					return nil, err
				}
				removed := pc.fw.RemoveRule(peerDID)
				return map[string]interface{}{
					"status":  "removed",
					"count":   removed,
					"message": fmt.Sprintf("Removed %d rules for %s", removed, peerDID),
				}, nil
			},
		},
		{
			Name:        "p2p_price_query",
			Description: "Query pricing for a specific tool on a remote peer before invoking it",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "p2p",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"peer_did":  map[string]interface{}{"type": "string", "description": "The remote peer's DID"},
					"tool_name": map[string]interface{}{"type": "string", "description": "The tool to query pricing for"},
				},
				"required": []string{"peer_did", "tool_name"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				peerDID, err := toolparam.RequireString(params, "peer_did")
				if err != nil {
					return nil, err
				}
				toolName, err := toolparam.RequireString(params, "tool_name")
				if err != nil {
					return nil, err
				}

				sess := pc.sessions.Get(peerDID)
				if sess == nil {
					return nil, fmt.Errorf("no active session for peer %s — connect first", peerDID)
				}

				did, err := identity.ParseDID(peerDID)
				if err != nil {
					return nil, fmt.Errorf("parse peer DID: %w", err)
				}

				remoteAgent := protocol.NewRemoteAgent(protocol.RemoteAgentConfig{
					Name:         "peer-" + peerDID[:16],
					DID:          peerDID,
					PeerID:       did.PeerID,
					SessionToken: sess.Token,
					Host:         pc.node.Host(),
					Logger:       logger(),
				})

				quote, err := remoteAgent.QueryPrice(ctx, toolName)
				if err != nil {
					return nil, fmt.Errorf("price query: %w", err)
				}

				return map[string]interface{}{
					"toolName":     quote.ToolName,
					"price":        quote.Price,
					"currency":     quote.Currency,
					"usdcContract": quote.USDCContract,
					"chainId":      quote.ChainID,
					"sellerAddr":   quote.SellerAddr,
					"quoteExpiry":  quote.QuoteExpiry,
					"isFree":       quote.IsFree,
				}, nil
			},
		},
		{
			Name:        "p2p_reputation",
			Description: "Check a peer's trust score and exchange history",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "p2p",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"peer_did": map[string]interface{}{"type": "string", "description": "The peer's DID to check reputation for"},
				},
				"required": []string{"peer_did"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				peerDID, err := toolparam.RequireString(params, "peer_did")
				if err != nil {
					return nil, err
				}

				if pc.reputation == nil {
					return nil, fmt.Errorf("reputation system not available (requires database)")
				}

				details, err := pc.reputation.GetDetails(ctx, peerDID)
				if err != nil {
					return nil, fmt.Errorf("get reputation: %w", err)
				}

				if details == nil {
					return map[string]interface{}{
						"peerDID":   peerDID,
						"score":     0.0,
						"isTrusted": true,
						"message":   "new peer — no reputation record",
					}, nil
				}

				return map[string]interface{}{
					"peerDID":             details.PeerDID,
					"trustScore":          details.TrustScore,
					"isTrusted":           details.TrustScore >= 0.3,
					"successfulExchanges": details.SuccessfulExchanges,
					"failedExchanges":     details.FailedExchanges,
					"timeoutCount":        details.TimeoutCount,
					"firstSeen":           details.FirstSeen.Format(time.RFC3339),
					"lastInteraction":     details.LastInteraction.Format(time.RFC3339),
				}, nil
			},
		},
		{
			Name:        "p2p_discover",
			Description: "Discover peers by capability or tags",
			SafetyLevel: agent.SafetyLevelSafe,
			Capability: agent.ToolCapability{
				Category:        "p2p",
				Activity:        agent.ActivityQuery,
				ReadOnly:        true,
				ConcurrencySafe: true,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"capability": map[string]interface{}{"type": "string", "description": "Capability to search for"},
				},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				capability := toolparam.OptionalString(params, "capability", "")

				if pc.gossip == nil {
					return map[string]interface{}{"peers": []interface{}{}, "count": 0, "message": "gossip not enabled"}, nil
				}

				var cards []*discovery.GossipCard
				if capability != "" {
					cards = pc.gossip.FindByCapability(capability)
				} else {
					cards = pc.gossip.KnownPeers()
				}

				peers := make([]map[string]interface{}, 0, len(cards))
				for _, c := range cards {
					peers = append(peers, map[string]interface{}{
						"name":         c.Name,
						"did":          c.DID,
						"capabilities": c.Capabilities,
						"pricing":      c.Pricing,
						"peerID":       c.PeerID,
						"timestamp":    c.Timestamp.Format(time.RFC3339),
					})
				}
				return map[string]interface{}{"peers": peers, "count": len(peers)}, nil
			},
		},
	}
}

// buildP2PPaymentTool creates the p2p_pay tool for peer-to-peer USDC payments.
func buildP2PPaymentTool(p2pc *p2pComponents, pc *paymentComponents) []*agent.Tool {
	if pc == nil || pc.service == nil {
		return nil
	}

	return []*agent.Tool{
		{
			Name:        "p2p_pay",
			Description: "Send USDC payment to a connected peer for their services",
			SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Category: "p2p",
				Activity: agent.ActivityExecute,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"peer_did": map[string]interface{}{"type": "string", "description": "The recipient peer's DID"},
					"amount":   map[string]interface{}{"type": "string", "description": "Amount in USDC (e.g., '0.50')"},
					"memo":     map[string]interface{}{"type": "string", "description": "Payment memo/reason"},
				},
				"required": []string{"peer_did", "amount"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				peerDID, err := toolparam.RequireString(params, "peer_did")
				if err != nil {
					return nil, err
				}
				amount, err := toolparam.RequireString(params, "amount")
				if err != nil {
					return nil, err
				}
				memo := toolparam.OptionalString(params, "memo", "")

				// Verify session exists for this peer.
				sess := p2pc.sessions.Get(peerDID)
				if sess == nil {
					return nil, fmt.Errorf("no active session for peer %s — connect first", peerDID)
				}

				// Get the peer's wallet address from their DID.
				did, err := identity.ParseDID(peerDID)
				if err != nil {
					return nil, fmt.Errorf("parse peer DID: %w", err)
				}

				// Derive Ethereum address from compressed public key.
				recipientAddr := fmt.Sprintf("0x%x", did.PublicKey[:20])

				if memo == "" {
					memo = "P2P payment"
				}

				sessionKey := session.SessionKeyFromContext(ctx)
				receipt, err := pc.service.Send(ctx, payment.PaymentRequest{
					To:         recipientAddr,
					Amount:     amount,
					Purpose:    memo,
					SessionKey: sessionKey,
				})
				if err != nil {
					return nil, fmt.Errorf("send payment: %w", err)
				}

				return map[string]interface{}{
					"status":    receipt.Status,
					"txHash":    receipt.TxHash,
					"from":      receipt.From,
					"to":        receipt.To,
					"peerDID":   peerDID,
					"amount":    receipt.Amount,
					"currency":  wallet.CurrencyUSDC,
					"chainId":   receipt.ChainID,
					"memo":      memo,
					"timestamp": receipt.Timestamp.Format(time.RFC3339),
				}, nil
			},
		},
	}
}

// authToMap serializes an eip3009.Authorization into the map format expected
// by the seller-side paygate parseAuthorization().
func authToMap(auth *eip3009.Authorization) map[string]interface{} {
	return map[string]interface{}{
		"from":        auth.From.Hex(),
		"to":          auth.To.Hex(),
		"value":       auth.Value.String(),
		"validAfter":  auth.ValidAfter.String(),
		"validBefore": auth.ValidBefore.String(),
		"nonce":       "0x" + hex.EncodeToString(auth.Nonce[:]),
		"v":           float64(auth.V),
		"r":           "0x" + hex.EncodeToString(auth.R[:]),
		"s":           "0x" + hex.EncodeToString(auth.S[:]),
	}
}

// paidInvokeDefaultDeadline is the EIP-3009 authorization validity window.
const paidInvokeDefaultDeadline = 10 * time.Minute

// buildP2PPaidInvokeTool creates the p2p_invoke_paid tool that automates
// buyer-side paid tool invocation: price query → spending check → EIP-3009
// signing → remote paid invoke.
func buildP2PPaidInvokeTool(p2pc *p2pComponents, pc *paymentComponents) []*agent.Tool {
	if pc == nil || pc.wallet == nil || pc.limiter == nil {
		return nil
	}

	usdcAddr, err := contracts.LookupUSDC(pc.chainID)
	if err != nil {
		logger().Warnw("p2p_invoke_paid: USDC contract lookup failed, skipping", "chainID", pc.chainID, "error", err)
		return nil
	}

	return []*agent.Tool{
		{
			Name:        "p2p_invoke_paid",
			Description: "Invoke a tool on a remote peer with automatic payment: queries price, checks spending limits, signs EIP-3009 authorization, and executes the paid call",
			SafetyLevel: agent.SafetyLevelDangerous,
			Capability: agent.ToolCapability{
				Category: "p2p",
				Activity: agent.ActivityExecute,
			},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"peer_did":  map[string]interface{}{"type": "string", "description": "The remote peer's DID"},
					"tool_name": map[string]interface{}{"type": "string", "description": "The tool to invoke on the remote agent"},
					"params":    map[string]interface{}{"type": "string", "description": "JSON string of parameters for the tool"},
				},
				"required": []string{"peer_did", "tool_name"},
			},
			Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
				peerDID, err := toolparam.RequireString(params, "peer_did")
				if err != nil {
					return nil, err
				}
				toolName, err := toolparam.RequireString(params, "tool_name")
				if err != nil {
					return nil, err
				}
				paramStr := toolparam.OptionalString(params, "params", "")

				// 1. Verify active session.
				sess := p2pc.sessions.Get(peerDID)
				if sess == nil {
					return nil, fmt.Errorf("no active session for peer %s — connect first", peerDID)
				}

				did, err := identity.ParseDID(peerDID)
				if err != nil {
					return nil, fmt.Errorf("parse peer DID: %w", err)
				}

				var toolParams map[string]interface{}
				if paramStr != "" {
					if err := json.Unmarshal([]byte(paramStr), &toolParams); err != nil {
						return nil, fmt.Errorf("parse params JSON: %w", err)
					}
				}
				if toolParams == nil {
					toolParams = map[string]interface{}{}
				}

				remoteAgent := protocol.NewRemoteAgent(protocol.RemoteAgentConfig{
					Name:         "peer-" + peerDID[:16],
					DID:          peerDID,
					PeerID:       did.PeerID,
					SessionToken: sess.Token,
					Host:         p2pc.node.Host(),
					Logger:       logger(),
				})

				// 2. Query price.
				quote, err := remoteAgent.QueryPrice(ctx, toolName)
				if err != nil {
					return nil, fmt.Errorf("price query: %w", err)
				}

				// 3. Free tool → invoke directly.
				if quote.IsFree {
					result, err := remoteAgent.InvokeTool(ctx, toolName, toolParams)
					if err != nil {
						return nil, fmt.Errorf("invoke free tool: %w", err)
					}
					return map[string]interface{}{
						"status": "ok",
						"paid":   false,
						"result": result,
					}, nil
				}

				// 4. Paid tool → build and sign EIP-3009 authorization.
				amount, err := wallet.ParseUSDC(quote.Price)
				if err != nil {
					return nil, fmt.Errorf("parse price %q: %w", quote.Price, err)
				}

				// 4a. Check spending limits.
				if err := pc.limiter.Check(ctx, amount); err != nil {
					return nil, fmt.Errorf("spending limit: %w", err)
				}

				// 4b. Check auto-approval threshold.
				autoApproved, err := pc.limiter.IsAutoApprovable(ctx, amount)
				if err != nil {
					return nil, fmt.Errorf("auto-approve check: %w", err)
				}
				if !autoApproved {
					return map[string]interface{}{
						"status":   "approval_required",
						"toolName": toolName,
						"price":    quote.Price,
						"currency": quote.Currency,
						"message":  fmt.Sprintf("Payment of %s %s requires explicit approval", quote.Price, quote.Currency),
					}, nil
				}

				// 4c. Build unsigned authorization.
				buyerAddr, err := pc.wallet.Address(ctx)
				if err != nil {
					return nil, fmt.Errorf("get wallet address: %w", err)
				}

				sellerAddr := common.HexToAddress(quote.SellerAddr)
				deadline := time.Now().Add(paidInvokeDefaultDeadline)

				unsigned := eip3009.NewUnsigned(
					common.HexToAddress(buyerAddr),
					sellerAddr,
					amount,
					deadline,
				)

				// 4d. Sign the authorization.
				signed, err := eip3009.Sign(ctx, pc.wallet, unsigned, pc.chainID, usdcAddr)
				if err != nil {
					return nil, fmt.Errorf("sign EIP-3009 authorization: %w", err)
				}

				// 4e. Invoke the paid tool.
				authMap := authToMap(signed)
				resp, err := remoteAgent.InvokeToolPaid(ctx, toolName, toolParams, authMap)
				if err != nil {
					return nil, fmt.Errorf("paid tool invoke: %w", err)
				}

				// 5. Handle response status.
				switch resp.Status {
				case protocol.ResponseStatusOK:
					// Record spending after successful invocation.
					if recordErr := pc.limiter.Record(ctx, amount); recordErr != nil {
						logger().Warnw("record spending after paid invoke", "error", recordErr)
					}
					return map[string]interface{}{
						"status":   "ok",
						"paid":     true,
						"price":    quote.Price,
						"currency": quote.Currency,
						"result":   resp.Result,
					}, nil

				case protocol.ResponseStatusPaymentRequired:
					return map[string]interface{}{
						"status":  "payment_required",
						"message": "seller rejected payment — authorization may be insufficient or expired",
						"detail":  resp.Result,
					}, nil

				default:
					errMsg := resp.Error
					if errMsg == "" {
						errMsg = "remote tool error"
					}
					return nil, fmt.Errorf("remote %s: %s", toolName, errMsg)
				}
			},
		},
	}
}
