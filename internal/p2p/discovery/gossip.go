// Package discovery implements gossip-based agent card propagation and peer discovery.
package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/zap"
)

// TopicAgentCard is the GossipSub topic for agent card propagation.
const TopicAgentCard = "/lango/agentcard/1.0.0"

// GossipCard is an agent card propagated via GossipSub.
type GossipCard struct {
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	DID          string       `json:"did,omitempty"`
	Multiaddrs   []string     `json:"multiaddrs,omitempty"`
	Capabilities []string     `json:"capabilities,omitempty"`
	Pricing      *PricingInfo `json:"pricing,omitempty"`
	ZKCredentials []ZKCredential `json:"zkCredentials,omitempty"`
	PeerID       string       `json:"peerId"`
	Timestamp    time.Time    `json:"timestamp"`
}

// PricingInfo describes the pricing for an agent's services.
type PricingInfo struct {
	Currency    string            `json:"currency"`    // e.g. "USDC"
	PerQuery    string            `json:"perQuery"`    // per-query price
	PerMinute   string            `json:"perMinute"`   // per-minute price
	ToolPrices  map[string]string `json:"toolPrices"`  // per-tool pricing
}

// ZKCredential is a zero-knowledge proof of agent capability.
type ZKCredential struct {
	CapabilityID string    `json:"capabilityId"`
	Proof        []byte    `json:"proof"`
	IssuedAt     time.Time `json:"issuedAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

// ZKCredentialVerifier verifies a ZK credential proof.
type ZKCredentialVerifier func(cred *ZKCredential) (bool, error)

// GossipService manages agent card propagation via GossipSub.
type GossipService struct {
	host      host.Host
	ps        *pubsub.PubSub
	topic     *pubsub.Topic
	sub       *pubsub.Subscription
	localCard *GossipCard
	interval  time.Duration
	verifier  ZKCredentialVerifier

	mu     sync.RWMutex
	peers  map[string]*GossipCard // keyed by DID
	cancel context.CancelFunc
	logger *zap.SugaredLogger
}

// GossipConfig configures the gossip service.
type GossipConfig struct {
	Host      host.Host
	LocalCard *GossipCard
	Interval  time.Duration
	Verifier  ZKCredentialVerifier
	Logger    *zap.SugaredLogger
}

// NewGossipService creates a new gossip-based discovery service.
func NewGossipService(cfg GossipConfig) (*GossipService, error) {
	ps, err := pubsub.NewGossipSub(context.Background(), cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("create gossipsub: %w", err)
	}

	topic, err := ps.Join(TopicAgentCard)
	if err != nil {
		return nil, fmt.Errorf("join topic %s: %w", TopicAgentCard, err)
	}

	sub, err := topic.Subscribe()
	if err != nil {
		return nil, fmt.Errorf("subscribe to %s: %w", TopicAgentCard, err)
	}

	return &GossipService{
		host:      cfg.Host,
		ps:        ps,
		topic:     topic,
		sub:       sub,
		localCard: cfg.LocalCard,
		interval:  cfg.Interval,
		verifier:  cfg.Verifier,
		peers:     make(map[string]*GossipCard),
		logger:    cfg.Logger,
	}, nil
}

// Start begins periodic card publication and message processing.
func (g *GossipService) Start(wg *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(context.Background())
	g.cancel = cancel

	// Publisher goroutine.
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.publishLoop(ctx)
	}()

	// Subscriber goroutine.
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.subscribeLoop(ctx)
	}()

	g.logger.Infow("gossip service started", "topic", TopicAgentCard, "interval", g.interval)
}

// Stop halts the gossip service.
func (g *GossipService) Stop() {
	if g.cancel != nil {
		g.cancel()
	}
	g.sub.Cancel()
	g.topic.Close()
	g.logger.Info("gossip service stopped")
}

// KnownPeers returns all known peer agent cards.
func (g *GossipService) KnownPeers() []*GossipCard {
	g.mu.RLock()
	defer g.mu.RUnlock()

	cards := make([]*GossipCard, 0, len(g.peers))
	for _, card := range g.peers {
		cards = append(cards, card)
	}
	return cards
}

// FindByCapability returns peers that advertise the given capability.
func (g *GossipService) FindByCapability(capability string) []*GossipCard {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var matches []*GossipCard
	for _, card := range g.peers {
		for _, cap := range card.Capabilities {
			if cap == capability {
				matches = append(matches, card)
				break
			}
		}
	}
	return matches
}

// FindByDID returns a peer by DID.
func (g *GossipService) FindByDID(did string) *GossipCard {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.peers[did]
}

// publishLoop periodically publishes the local agent card.
func (g *GossipService) publishLoop(ctx context.Context) {
	ticker := time.NewTicker(g.interval)
	defer ticker.Stop()

	// Publish immediately on start.
	g.publishCard(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			g.publishCard(ctx)
		}
	}
}

// publishCard publishes the local agent card to the gossip topic.
func (g *GossipService) publishCard(ctx context.Context) {
	if g.localCard == nil {
		return
	}

	g.localCard.Timestamp = time.Now()

	data, err := json.Marshal(g.localCard)
	if err != nil {
		g.logger.Warnw("marshal agent card", "error", err)
		return
	}

	if err := g.topic.Publish(ctx, data); err != nil {
		g.logger.Debugw("publish agent card", "error", err)
	}
}

// subscribeLoop processes incoming agent card messages.
func (g *GossipService) subscribeLoop(ctx context.Context) {
	for {
		msg, err := g.sub.Next(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			g.logger.Warnw("gossip subscription", "error", err)
			continue
		}

		// Skip own messages.
		if msg.ReceivedFrom == g.host.ID() {
			continue
		}

		g.handleMessage(msg)
	}
}

// handleMessage processes a received gossip message.
func (g *GossipService) handleMessage(msg *pubsub.Message) {
	var card GossipCard
	if err := json.Unmarshal(msg.Data, &card); err != nil {
		g.logger.Debugw("unmarshal gossip card", "error", err, "from", msg.ReceivedFrom)
		return
	}

	if card.DID == "" {
		return
	}

	// Verify ZK credentials if verifier is available.
	if g.verifier != nil {
		for _, cred := range card.ZKCredentials {
			if cred.ExpiresAt.Before(time.Now()) {
				g.logger.Debugw("expired ZK credential", "did", card.DID, "capability", cred.CapabilityID)
				continue
			}
			valid, err := g.verifier(&cred)
			if err != nil || !valid {
				g.logger.Warnw("invalid ZK credential, discarding card",
					"did", card.DID,
					"capability", cred.CapabilityID,
					"error", err,
				)
				return // Discard the entire card if any credential is invalid.
			}
		}
	}

	// Store/update peer card.
	g.mu.Lock()
	existing, ok := g.peers[card.DID]
	if !ok || card.Timestamp.After(existing.Timestamp) {
		g.peers[card.DID] = &card
		g.logger.Debugw("peer card updated",
			"did", card.DID,
			"name", card.Name,
			"capabilities", card.Capabilities,
		)
	}
	g.mu.Unlock()
}

// PeerIDFromString parses a peer ID string.
func PeerIDFromString(s string) (peer.ID, error) {
	return peer.Decode(s)
}
