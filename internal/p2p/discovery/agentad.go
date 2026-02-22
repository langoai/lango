package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"go.uber.org/zap"
)

// AgentAd is a structured service advertisement (Context Flyer).
type AgentAd struct {
	DID          string            `json:"did"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Tags         []string          `json:"tags"`
	Capabilities []string          `json:"capabilities,omitempty"`
	Pricing      *PricingInfo      `json:"pricing,omitempty"`
	ZKCredentials []ZKCredential   `json:"zkCredentials,omitempty"`
	Multiaddrs   []string          `json:"multiaddrs,omitempty"`
	PeerID       string            `json:"peerId"`
	Timestamp    time.Time         `json:"timestamp"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// AdService manages agent advertisement and discovery via DHT provider records.
type AdService struct {
	dht      *dht.IpfsDHT
	localAd  *AgentAd
	verifier ZKCredentialVerifier
	mu       sync.RWMutex
	ads      map[string]*AgentAd // keyed by DID
	logger   *zap.SugaredLogger
}

// AdServiceConfig configures the AdService.
type AdServiceConfig struct {
	DHT      *dht.IpfsDHT
	LocalAd  *AgentAd
	Verifier ZKCredentialVerifier
	Logger   *zap.SugaredLogger
}

// NewAdService creates a new agent advertisement service.
func NewAdService(cfg AdServiceConfig) *AdService {
	return &AdService{
		dht:      cfg.DHT,
		localAd:  cfg.LocalAd,
		verifier: cfg.Verifier,
		ads:      make(map[string]*AgentAd),
		logger:   cfg.Logger,
	}
}

// Advertise publishes the local agent ad to the DHT.
func (s *AdService) Advertise(ctx context.Context) error {
	if s.localAd == nil {
		return nil
	}

	s.localAd.Timestamp = time.Now()

	data, err := json.Marshal(s.localAd)
	if err != nil {
		return fmt.Errorf("marshal agent ad: %w", err)
	}

	// Store as a DHT value keyed by the agent's DID.
	key := "/lango/agentad/" + s.localAd.DID
	if err := s.dht.PutValue(ctx, key, data); err != nil {
		return fmt.Errorf("put agent ad to DHT: %w", err)
	}

	s.logger.Debugw("agent ad advertised", "did", s.localAd.DID, "tags", s.localAd.Tags)
	return nil
}

// Discover searches for agent ads matching the given tags.
func (s *AdService) Discover(ctx context.Context, tags []string) ([]*AgentAd, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(tags) == 0 {
		// Return all known ads.
		ads := make([]*AgentAd, 0, len(s.ads))
		for _, ad := range s.ads {
			ads = append(ads, ad)
		}
		return ads, nil
	}

	// Filter by tags.
	var matches []*AgentAd
	for _, ad := range s.ads {
		if matchesTags(ad.Tags, tags) {
			matches = append(matches, ad)
		}
	}
	return matches, nil
}

// DiscoverByCapability returns ads matching a specific capability.
func (s *AdService) DiscoverByCapability(ctx context.Context, capability string) []*AgentAd {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matches []*AgentAd
	for _, ad := range s.ads {
		for _, cap := range ad.Capabilities {
			if cap == capability {
				matches = append(matches, ad)
				break
			}
		}
	}
	return matches
}

// StoreAd stores a discovered agent ad after ZK credential verification.
func (s *AdService) StoreAd(ad *AgentAd) error {
	if ad.DID == "" {
		return fmt.Errorf("agent ad missing DID")
	}

	// Verify ZK credentials if verifier is available.
	if s.verifier != nil {
		for _, cred := range ad.ZKCredentials {
			if cred.ExpiresAt.Before(time.Now()) {
				s.logger.Debugw("expired ZK credential in ad", "did", ad.DID, "capability", cred.CapabilityID)
				continue
			}
			valid, err := s.verifier(&cred)
			if err != nil || !valid {
				return fmt.Errorf("invalid ZK credential in ad for %s: capability %s", ad.DID, cred.CapabilityID)
			}
		}
	}

	s.mu.Lock()
	existing, ok := s.ads[ad.DID]
	if !ok || ad.Timestamp.After(existing.Timestamp) {
		s.ads[ad.DID] = ad
	}
	s.mu.Unlock()

	s.logger.Debugw("agent ad stored", "did", ad.DID, "name", ad.Name, "tags", ad.Tags)
	return nil
}

// matchesTags returns true if the ad tags contain any of the requested tags.
func matchesTags(adTags, requestedTags []string) bool {
	tagSet := make(map[string]struct{}, len(adTags))
	for _, t := range adTags {
		tagSet[t] = struct{}{}
	}
	for _, t := range requestedTags {
		if _, ok := tagSet[t]; ok {
			return true
		}
	}
	return false
}
