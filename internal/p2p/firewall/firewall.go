// Package firewall implements the Knowledge Firewall for P2P queries.
// Default policy is deny-all — explicit rules must be added to allow access.
package firewall

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// ACLRule defines an access control rule.
type ACLRule struct {
	// PeerDID is the peer this rule applies to ("*" for all peers).
	PeerDID string `json:"peerDid"`

	// Action is "allow" or "deny".
	Action string `json:"action"`

	// Tools lists tool name patterns (supports * wildcard).
	Tools []string `json:"tools"`

	// RateLimit is max requests per minute (0 = unlimited).
	RateLimit int `json:"rateLimit"`
}

// AttestationResult holds a structured ZK attestation proof from the prover.
type AttestationResult struct {
	Proof        []byte
	PublicInputs []byte
	CircuitID    string
	Scheme       string
}

// ZKAttestFunc generates a ZK attestation proof for a response.
type ZKAttestFunc func(responseHash, agentDIDHash []byte) (*AttestationResult, error)

// ReputationChecker returns a trust score for a peer DID.
type ReputationChecker func(ctx context.Context, peerDID string) (float64, error)

// Firewall enforces access control and response sanitization for P2P queries.
type Firewall struct {
	rules           []ACLRule
	mu              sync.RWMutex
	limiters        map[string]*rate.Limiter // per-peer rate limiters
	attestFunc      ZKAttestFunc
	ownerShield     *OwnerShield
	reputationCheck ReputationChecker
	minTrustScore   float64
	logger          *zap.SugaredLogger
}

// New creates a new Firewall with deny-all default policy.
func New(rules []ACLRule, logger *zap.SugaredLogger) *Firewall {
	f := &Firewall{
		rules:    make([]ACLRule, len(rules)),
		limiters: make(map[string]*rate.Limiter),
		logger:   logger,
	}
	copy(f.rules, rules)

	// Initialize per-peer rate limiters from rules.
	for _, r := range rules {
		if r.RateLimit > 0 && r.PeerDID != "" {
			f.limiters[r.PeerDID] = rate.NewLimiter(rate.Every(time.Minute/time.Duration(r.RateLimit)), r.RateLimit)
		}
	}

	return f
}

// SetZKAttestFunc sets the ZK attestation function for response signing.
func (f *Firewall) SetZKAttestFunc(fn ZKAttestFunc) {
	f.mu.Lock()
	f.attestFunc = fn
	f.mu.Unlock()
}

// SetOwnerShield sets the owner data protection shield.
func (f *Firewall) SetOwnerShield(shield *OwnerShield) {
	f.mu.Lock()
	f.ownerShield = shield
	f.mu.Unlock()
}

// SetReputationChecker sets the reputation checker and minimum trust score.
func (f *Firewall) SetReputationChecker(fn ReputationChecker, minScore float64) {
	f.mu.Lock()
	f.reputationCheck = fn
	f.minTrustScore = minScore
	f.mu.Unlock()
}

// FilterQuery checks if a query from the given peer is allowed.
func (f *Firewall) FilterQuery(ctx context.Context, peerDID, toolName string) error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	// Check rate limit first.
	if limiter, ok := f.limiters[peerDID]; ok {
		if !limiter.Allow() {
			return fmt.Errorf("rate limit exceeded for peer %s", peerDID)
		}
	}
	// Also check wildcard rate limiter.
	if limiter, ok := f.limiters["*"]; ok {
		if !limiter.Allow() {
			return fmt.Errorf("global rate limit exceeded")
		}
	}

	// Check reputation score.
	if f.reputationCheck != nil {
		score, err := f.reputationCheck(ctx, peerDID)
		if err != nil {
			f.logger.Warnw("reputation check error", "peerDID", peerDID, "error", err)
			// Don't block on reputation errors, continue to ACL.
		} else if score > 0 && score < f.minTrustScore {
			// score == 0 means new peer, allow through (they start fresh).
			return fmt.Errorf("peer %s reputation %.2f below minimum %.2f", peerDID, score, f.minTrustScore)
		}
	}

	// Check ACL rules. Default is deny-all.
	allowed := false
	for _, rule := range f.rules {
		if !matchesPeer(rule.PeerDID, peerDID) {
			continue
		}
		if !matchesTool(rule.Tools, toolName) {
			continue
		}

		switch rule.Action {
		case "allow":
			allowed = true
		case "deny":
			return fmt.Errorf("query denied by firewall rule for peer %s, tool %s", peerDID, toolName)
		}
	}

	if !allowed {
		return fmt.Errorf("query denied: no matching allow rule for peer %s, tool %s", peerDID, toolName)
	}

	return nil
}

// SanitizeResponse removes sensitive internal data from a response.
func (f *Firewall) SanitizeResponse(response map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{}, len(response))

	for k, v := range response {
		// Remove internal fields that should never be exposed.
		if isSensitiveKey(k) {
			continue
		}

		switch val := v.(type) {
		case string:
			sanitized[k] = sanitizeString(val)
		case map[string]interface{}:
			sanitized[k] = f.SanitizeResponse(val)
		default:
			sanitized[k] = v
		}
	}

	// Apply owner shield if configured.
	if f.ownerShield != nil {
		var blocked []string
		sanitized, blocked = f.ownerShield.ScanAndRedact(sanitized)
		if len(blocked) > 0 {
			f.logger.Infow("owner data redacted from P2P response", "fields", blocked)
		}
	}

	return sanitized
}

// AttestResponse generates a ZK attestation proof for a response.
func (f *Firewall) AttestResponse(responseHash, agentDIDHash []byte) (*AttestationResult, error) {
	f.mu.RLock()
	fn := f.attestFunc
	f.mu.RUnlock()

	if fn == nil {
		return nil, nil // Attestation not configured.
	}

	return fn(responseHash, agentDIDHash)
}

// AddRule adds a new ACL rule.
func (f *Firewall) AddRule(rule ACLRule) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.rules = append(f.rules, rule)

	if rule.RateLimit > 0 && rule.PeerDID != "" {
		f.limiters[rule.PeerDID] = rate.NewLimiter(rate.Every(time.Minute/time.Duration(rule.RateLimit)), rule.RateLimit)
	}

	f.logger.Infow("firewall rule added",
		"peerDID", rule.PeerDID,
		"action", rule.Action,
		"tools", rule.Tools,
	)
}

// RemoveRule removes ACL rules matching the peer DID.
func (f *Firewall) RemoveRule(peerDID string) int {
	f.mu.Lock()
	defer f.mu.Unlock()

	var kept []ACLRule
	removed := 0
	for _, r := range f.rules {
		if r.PeerDID == peerDID {
			removed++
			continue
		}
		kept = append(kept, r)
	}
	f.rules = kept
	delete(f.limiters, peerDID)

	f.logger.Infow("firewall rules removed", "peerDID", peerDID, "count", removed)
	return removed
}

// Rules returns a copy of current rules.
func (f *Firewall) Rules() []ACLRule {
	f.mu.RLock()
	defer f.mu.RUnlock()

	rules := make([]ACLRule, len(f.rules))
	copy(rules, f.rules)
	return rules
}

// matchesPeer checks if a rule peer pattern matches the given peer DID.
func matchesPeer(pattern, peerDID string) bool {
	if pattern == "*" {
		return true
	}
	return pattern == peerDID
}

// matchesTool checks if any tool pattern in the rule matches the tool name.
func matchesTool(patterns []string, toolName string) bool {
	if len(patterns) == 0 {
		return true // No tool filter means all tools.
	}
	for _, p := range patterns {
		if p == "*" {
			return true
		}
		if strings.HasSuffix(p, "*") {
			if strings.HasPrefix(toolName, strings.TrimSuffix(p, "*")) {
				return true
			}
		}
		if p == toolName {
			return true
		}
	}
	return false
}

// sensitiveKeyPatterns are field names that should be stripped from responses.
var sensitiveKeyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)^(db_?path|file_?path|internal_?id|_internal)$`),
	regexp.MustCompile(`(?i)password|secret|private_?key|token`),
}

// isSensitiveKey checks if a response field name should be stripped.
func isSensitiveKey(key string) bool {
	for _, re := range sensitiveKeyPatterns {
		if re.MatchString(key) {
			return true
		}
	}
	return false
}

// sanitizeString removes file paths and internal references from string values.
func sanitizeString(s string) string {
	// Remove absolute file paths.
	pathPattern := regexp.MustCompile(`(?:/[a-zA-Z0-9._-]+){3,}`)
	s = pathPattern.ReplaceAllString(s, "[path-redacted]")

	return s
}
