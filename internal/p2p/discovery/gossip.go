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

	"github.com/langoai/lango/internal/p2p/identity"
	"github.com/langoai/lango/internal/security"
)

// TopicAgentCard is the GossipSub topic for agent card propagation.
const TopicAgentCard = "/lango/agentcard/1.0.0"

// GossipCard is an agent card propagated via GossipSub.
type GossipCard struct {
	Name               string          `json:"name"`
	Description        string          `json:"description"`
	DID                string          `json:"did,omitempty"`
	Multiaddrs         []string        `json:"multiaddrs,omitempty"`
	Capabilities       []string        `json:"capabilities,omitempty"`
	Pricing            *PricingInfo    `json:"pricing,omitempty"`
	ZKCredentials      []ZKCredential  `json:"zkCredentials,omitempty"`
	OntologyDigest     *OntologyDigest `json:"ontologyDigest,omitempty"`
	PeerID             string          `json:"peerId"`
	Timestamp          time.Time       `json:"timestamp"`
	Bundle             json.RawMessage `json:"bundle,omitempty"`             // v2: serialized IdentityBundle for DID resolution
	Signature          []byte          `json:"signature,omitempty"`          // classical signature over canonical payload
	SignatureAlgorithm string          `json:"signatureAlgorithm,omitempty"` // algorithm for classical signature
	PQSignerPublicKey  []byte          `json:"pqSignerPublicKey,omitempty"`  // embedded ML-DSA-65 pubkey for rotation-safe PQ verification
	PQSignature        []byte          `json:"pqSignature,omitempty"`        // ML-DSA-65 signature over canonical payload
	PQSignatureAlgorithm string        `json:"pqSignatureAlgorithm,omitempty"`
}

// PricingInfo describes the pricing for an agent's services.
type PricingInfo struct {
	Currency   string            `json:"currency"`   // e.g. "USDC"
	PerQuery   string            `json:"perQuery"`   // per-query price
	PerMinute  string            `json:"perMinute"`  // per-minute price
	ToolPrices map[string]string `json:"toolPrices"` // per-tool pricing
}

// OntologyDigest is a lightweight summary of an agent's ontology schema,
// enabling peers to discover ontology-capable agents and assess schema
// compatibility before initiating heavier exchange protocols.
type OntologyDigest struct {
	SchemaVersion  int      `json:"schemaVersion"`
	Digest         string   `json:"digest"`
	TypeCount      int      `json:"typeCount"`
	PredicateCount int      `json:"predicateCount"`
	TypeNames      []string `json:"typeNames,omitempty"`
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

// CardSigner signs gossip card content with a classical algorithm.
type CardSigner interface {
	Sign(ctx context.Context, payload []byte) ([]byte, error)
	Algorithm() string
}

// PQCardSigner is an optional interface for PQ dual-signing gossip cards.
type PQCardSigner interface {
	SignPQ(ctx context.Context, payload []byte) ([]byte, error)
	PQAlgorithm() string
	PQPublicKey() []byte
}

// CanonicalCardPayload returns the canonical JSON for signing.
// It clones the card, zeros ONLY Signature and PQSignature, and marshals
// everything else — including Bundle and PQSignerPublicKey.
func CanonicalCardPayload(card *GossipCard) ([]byte, error) {
	clone := *card
	clone.Signature = nil
	clone.PQSignature = nil
	return json.Marshal(clone)
}

// VerifyCardSignature verifies the classical + optional PQ signatures on a card.
// Returns nil if the card is unsigned (backward compat with legacy cards).
func VerifyCardSignature(card *GossipCard, classicalVerify func(pubkey, message, sig []byte) error) error {
	if len(card.Signature) == 0 {
		return nil // unsigned legacy card — accepted
	}
	payload, err := CanonicalCardPayload(card)
	if err != nil {
		return fmt.Errorf("canonical card payload: %w", err)
	}
	// Classical signature verification.
	// If bundle is absent (pre-upgrade peers), skip — backward compatibility.
	// If bundle is present, it MUST contain a valid signing key.
	if len(card.Bundle) > 0 {
		var bundle struct {
			SigningKey struct {
				PublicKey []byte `json:"public_key"`
			} `json:"signing_key"`
		}
		var pubkey []byte
		if err := json.Unmarshal(card.Bundle, &bundle); err == nil {
			pubkey = bundle.SigningKey.PublicKey
		}
		if len(pubkey) == 0 {
			return fmt.Errorf("signed card has bundle but no valid signing key")
		}
		if classicalVerify != nil {
			if err := classicalVerify(pubkey, payload, card.Signature); err != nil {
				return fmt.Errorf("verify card signature: %w", err)
			}
		}
	}
	// PQ verification uses embedded public key (self-contained, rotation-safe).
	// Skip PQ verification for bundle-less cards (pre-upgrade peers may have
	// signed PQ over a different canonical payload before SignatureAlgorithm
	// was populated).
	if len(card.PQSignature) > 0 && len(card.PQSignerPublicKey) > 0 && len(card.Bundle) > 0 {
		if err := security.VerifyMLDSA65(card.PQSignerPublicKey, payload, card.PQSignature); err != nil {
			return fmt.Errorf("verify card PQ signature: %w", err)
		}
	}
	// Verify card.DID matches the bundle's v2 DID to prevent impersonation.
	// Only ComputeDIDv2 is accepted — LegacyDID is a self-reported field
	// without Proofs.Legacy verification, so accepting it would allow
	// identity spoofing. Upgraded peers always use v2 DID in their cards.
	if len(card.Bundle) > 0 && card.DID != "" {
		var bundle identity.IdentityBundle
		if unmarshalErr := json.Unmarshal(card.Bundle, &bundle); unmarshalErr == nil {
			didV2, _ := identity.ComputeDIDv2(&bundle)
			if card.DID != didV2 {
				return fmt.Errorf("card DID %q does not match bundle v2 DID", card.DID)
			}
		}
	}
	return nil
}

// defaultMaxCredentialAge is the default maximum age for ZK credentials.
const defaultMaxCredentialAge = 24 * time.Hour

// GossipService manages agent card propagation via GossipSub.
type GossipService struct {
	host        host.Host
	ps          *pubsub.PubSub
	topic       *pubsub.Topic
	sub         *pubsub.Subscription
	localCard   *GossipCard
	interval    time.Duration
	verifier    ZKCredentialVerifier
	cardSigner      CardSigner              // optional: for signing published cards
	pqSigner        PQCardSigner            // optional: for PQ dual-signing
	classicalVerify CardSignatureVerifyFunc  // optional: verify received card signatures

	mu     sync.RWMutex
	peers  map[string]*GossipCard // keyed by DID
	cancel context.CancelFunc
	logger *zap.SugaredLogger

	revokedMu        sync.RWMutex
	revokedDIDs      map[string]time.Time // DID → revocation time
	maxCredentialAge time.Duration
}

// CardSignatureVerifyFunc verifies a card's classical signature.
type CardSignatureVerifyFunc func(pubkey, message, sig []byte) error

// GossipConfig configures the gossip service.
type GossipConfig struct {
	Host            host.Host
	PubSub          *pubsub.PubSub // optional pre-created PubSub instance
	LocalCard       *GossipCard
	Interval        time.Duration
	Verifier        ZKCredentialVerifier
	CardSigner      CardSigner            // optional: sign published cards
	PQCardSigner    PQCardSigner          // optional: PQ dual-sign
	ClassicalVerify CardSignatureVerifyFunc // optional: verify received card signatures
	Logger          *zap.SugaredLogger
}

// NewGossipService creates a new gossip-based discovery service.
func NewGossipService(cfg GossipConfig) (*GossipService, error) {
	ps := cfg.PubSub
	if ps == nil {
		var err error
		ps, err = pubsub.NewGossipSub(context.Background(), cfg.Host)
		if err != nil {
			return nil, fmt.Errorf("create gossipsub: %w", err)
		}
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
		host:             cfg.Host,
		ps:               ps,
		topic:            topic,
		sub:              sub,
		localCard:        cfg.LocalCard,
		interval:         cfg.Interval,
		verifier:         cfg.Verifier,
		cardSigner:       cfg.CardSigner,
		pqSigner:         cfg.PQCardSigner,
		classicalVerify:  cfg.ClassicalVerify,
		peers:            make(map[string]*GossipCard),
		logger:           cfg.Logger,
		revokedDIDs:      make(map[string]time.Time),
		maxCredentialAge: defaultMaxCredentialAge,
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

// RevokeDID marks a DID as revoked, preventing its credentials from being accepted.
func (g *GossipService) RevokeDID(did string) {
	g.revokedMu.Lock()
	g.revokedDIDs[did] = time.Now()
	g.revokedMu.Unlock()
	g.logger.Infow("DID revoked", "did", did)
}

// IsRevoked checks if a DID has been revoked.
func (g *GossipService) IsRevoked(did string) bool {
	g.revokedMu.RLock()
	_, revoked := g.revokedDIDs[did]
	g.revokedMu.RUnlock()
	return revoked
}

// SetMaxCredentialAge sets the maximum allowed age for ZK credentials.
func (g *GossipService) SetMaxCredentialAge(d time.Duration) {
	g.revokedMu.Lock()
	g.maxCredentialAge = d
	g.revokedMu.Unlock()
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

	// Sign the card before publishing.
	g.signCard(ctx, g.localCard)

	data, err := json.Marshal(g.localCard)
	if err != nil {
		g.logger.Warnw("marshal agent card", "error", err)
		return
	}

	if err := g.topic.Publish(ctx, data); err != nil {
		g.logger.Debugw("publish agent card", "error", err)
	}
}

// signCard applies classical and PQ signatures to a gossip card.
func (g *GossipService) signCard(ctx context.Context, card *GossipCard) {
	if g.cardSigner == nil {
		return
	}

	// Set SignatureAlgorithm BEFORE canonical payload computation.
	// CanonicalCardPayload includes SignatureAlgorithm, so it must be set
	// before signing to ensure sender and receiver hash the same JSON.
	card.SignatureAlgorithm = g.cardSigner.Algorithm()

	// PQ pubkey must be set before canonical payload computation
	// (included in classical signature's payload for trust chain binding).
	if g.pqSigner != nil {
		card.PQSignerPublicKey = g.pqSigner.PQPublicKey()
		card.PQSignatureAlgorithm = g.pqSigner.PQAlgorithm()
	}

	payload, err := CanonicalCardPayload(card)
	if err != nil {
		g.logger.Warnw("canonical card payload", "error", err)
		return
	}

	// Classical signature.
	sig, err := g.cardSigner.Sign(ctx, payload)
	if err != nil {
		g.logger.Warnw("sign gossip card", "error", err)
		return
	}
	card.Signature = sig

	// PQ signature (optional).
	if g.pqSigner != nil {
		pqSig, pqErr := g.pqSigner.SignPQ(ctx, payload)
		if pqErr != nil {
			g.logger.Warnw("PQ sign gossip card", "error", pqErr)
			return
		}
		card.PQSignature = pqSig
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

	// Reject cards from revoked DIDs.
	if g.IsRevoked(card.DID) {
		g.logger.Warnw("rejected card from revoked DID", "did", card.DID)
		return
	}

	// Verify ZK credentials if verifier is available.
	now := time.Now()
	if g.verifier != nil {
		for _, cred := range card.ZKCredentials {
			if cred.ExpiresAt.Before(now) {
				g.logger.Debugw("expired ZK credential",
					"did", card.DID, "capability", cred.CapabilityID)
				continue
			}

			// Check credential age against max allowed age.
			g.revokedMu.RLock()
			maxAge := g.maxCredentialAge
			g.revokedMu.RUnlock()
			if cred.IssuedAt.Add(maxAge).Before(now) {
				g.logger.Warnw("stale ZK credential exceeds max age",
					"did", card.DID,
					"capability", cred.CapabilityID,
					"issuedAt", cred.IssuedAt,
					"maxAge", maxAge,
				)
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

	// Verify card signature if present. Unsigned legacy cards are accepted.
	if err := VerifyCardSignature(&card, g.classicalVerify); err != nil {
		g.logger.Warnw("invalid card signature, discarding", "did", card.DID, "error", err)
		return
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
