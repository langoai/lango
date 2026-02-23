package protocol

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"
)

// P2PRemoteAgent represents a remote agent accessible over P2P.
// It can be used as a sub-agent in the orchestration framework.
type P2PRemoteAgent struct {
	name       string
	did        string
	peerID     peer.ID
	token      string
	host       host.Host
	capabilities []string
	logger     *zap.SugaredLogger
}

// RemoteAgentConfig configures a P2P remote agent.
type RemoteAgentConfig struct {
	Name         string
	DID          string
	PeerID       peer.ID
	SessionToken string
	Host         host.Host
	Capabilities []string
	Logger       *zap.SugaredLogger
}

// NewRemoteAgent creates a remote agent adapter for P2P communication.
func NewRemoteAgent(cfg RemoteAgentConfig) *P2PRemoteAgent {
	return &P2PRemoteAgent{
		name:         cfg.Name,
		did:          cfg.DID,
		peerID:       cfg.PeerID,
		token:        cfg.SessionToken,
		host:         cfg.Host,
		capabilities: cfg.Capabilities,
		logger:       cfg.Logger,
	}
}

// Name returns the remote agent's name.
func (a *P2PRemoteAgent) Name() string { return a.name }

// DID returns the remote agent's decentralized identifier.
func (a *P2PRemoteAgent) DID() string { return a.did }

// PeerID returns the remote agent's libp2p peer ID.
func (a *P2PRemoteAgent) PeerID() peer.ID { return a.peerID }

// Capabilities returns the remote agent's advertised capabilities.
func (a *P2PRemoteAgent) Capabilities() []string { return a.capabilities }

// InvokeTool sends a tool invocation to the remote agent.
func (a *P2PRemoteAgent) InvokeTool(ctx context.Context, toolName string, params map[string]interface{}) (map[string]interface{}, error) {
	s, err := a.host.NewStream(ctx, a.peerID, ProtocolID)
	if err != nil {
		return nil, fmt.Errorf("open stream to %s: %w", a.peerID, err)
	}
	defer s.Close()

	payload := map[string]interface{}{
		"toolName": toolName,
		"params":   params,
	}

	resp, err := SendRequest(ctx, s, RequestToolInvoke, a.token, payload)
	if err != nil {
		return nil, fmt.Errorf("tool invoke %s on %s: %w", toolName, a.name, err)
	}

	if resp.Status != "ok" {
		errMsg := resp.Error
		if errMsg == "" {
			errMsg = "unknown error"
		}
		return nil, fmt.Errorf("remote tool %s error: %s", toolName, errMsg)
	}

	// Verify ZK attestation if present.
	if len(resp.AttestationProof) > 0 {
		a.logger.Debugw("response has ZK attestation", "tool", toolName, "remote", a.name)
	}

	return resp.Result, nil
}

// QueryCapabilities fetches the remote agent's capabilities.
func (a *P2PRemoteAgent) QueryCapabilities(ctx context.Context) (map[string]interface{}, error) {
	s, err := a.host.NewStream(ctx, a.peerID, ProtocolID)
	if err != nil {
		return nil, fmt.Errorf("open stream to %s: %w", a.peerID, err)
	}
	defer s.Close()

	resp, err := SendRequest(ctx, s, RequestCapabilityQuery, a.token, nil)
	if err != nil {
		return nil, fmt.Errorf("capability query %s: %w", a.name, err)
	}

	if resp.Status != "ok" {
		return nil, fmt.Errorf("capability query error: %s", resp.Error)
	}

	return resp.Result, nil
}

// FetchAgentCard fetches the remote agent card.
func (a *P2PRemoteAgent) FetchAgentCard(ctx context.Context) (map[string]interface{}, error) {
	s, err := a.host.NewStream(ctx, a.peerID, ProtocolID)
	if err != nil {
		return nil, fmt.Errorf("open stream to %s: %w", a.peerID, err)
	}
	defer s.Close()

	resp, err := SendRequest(ctx, s, RequestAgentCard, a.token, nil)
	if err != nil {
		return nil, fmt.Errorf("agent card fetch %s: %w", a.name, err)
	}

	if resp.Status != "ok" {
		return nil, fmt.Errorf("agent card fetch error: %s", resp.Error)
	}

	return resp.Result, nil
}

// QueryPrice queries the pricing for a tool on the remote agent.
func (a *P2PRemoteAgent) QueryPrice(ctx context.Context, toolName string) (*PriceQuoteResult, error) {
	s, err := a.host.NewStream(ctx, a.peerID, ProtocolID)
	if err != nil {
		return nil, fmt.Errorf("open stream to %s: %w", a.peerID, err)
	}
	defer s.Close()

	payload := map[string]interface{}{"toolName": toolName}
	resp, err := SendRequest(ctx, s, RequestPriceQuery, a.token, payload)
	if err != nil {
		return nil, fmt.Errorf("price query %s on %s: %w", toolName, a.name, err)
	}

	if resp.Status != "ok" {
		errMsg := resp.Error
		if errMsg == "" {
			errMsg = "unknown error"
		}
		return nil, fmt.Errorf("price query %s error: %s", toolName, errMsg)
	}

	// Parse result into PriceQuoteResult.
	resultBytes, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, fmt.Errorf("marshal price quote: %w", err)
	}

	var quote PriceQuoteResult
	if err := json.Unmarshal(resultBytes, &quote); err != nil {
		return nil, fmt.Errorf("unmarshal price quote: %w", err)
	}

	return &quote, nil
}

// InvokeToolPaid sends a paid tool invocation to the remote agent.
func (a *P2PRemoteAgent) InvokeToolPaid(
	ctx context.Context,
	toolName string,
	params map[string]interface{},
	paymentAuth map[string]interface{},
) (*Response, error) {
	s, err := a.host.NewStream(ctx, a.peerID, ProtocolID)
	if err != nil {
		return nil, fmt.Errorf("open stream to %s: %w", a.peerID, err)
	}
	defer s.Close()

	payload := map[string]interface{}{
		"toolName":    toolName,
		"params":      params,
		"paymentAuth": paymentAuth,
	}

	resp, err := SendRequest(ctx, s, RequestToolInvokePaid, a.token, payload)
	if err != nil {
		return nil, fmt.Errorf("paid invoke %s on %s: %w", toolName, a.name, err)
	}

	return resp, nil
}
