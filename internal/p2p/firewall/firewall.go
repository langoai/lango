// Package firewall implements the Knowledge Firewall for P2P queries.
// Default policy is deny-all — explicit rules must be added to allow access.
package firewall

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// ACLAction identifies the action of an ACL rule.
type ACLAction string

const (
	// ACLActionAllow permits matching queries.
	ACLActionAllow ACLAction = "allow"

	// ACLActionDeny blocks matching queries.
	ACLActionDeny ACLAction = "deny"
)

// Valid reports whether a is a known ACL action.
func (a ACLAction) Valid() bool {
	switch a {
	case ACLActionAllow, ACLActionDeny:
		return true
	}
	return false
}

// WildcardAll matches all peers or all tools.
const WildcardAll = "*"

// Sentinel errors for firewall decisions.
var (
	ErrRateLimitExceeded       = errors.New("rate limit exceeded")
	ErrGlobalRateLimitExceeded = errors.New("global rate limit exceeded")
	ErrQueryDenied             = errors.New("query denied by firewall rule")
	ErrNoMatchingAllowRule     = errors.New("query denied: no matching allow rule")
)

// ACLRule defines an access control rule.
type ACLRule struct {
	// PeerDID is the peer this rule applies to (WildcardAll for all peers).
	PeerDID string `json:"peerDid"`

	// Action is ACLActionAllow or ACLActionDeny.
	Action ACLAction `json:"action"`

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

// ReputationAssessment summarizes the runtime trust-entry contract for a peer.
type ReputationAssessment struct {
	Score             float64
	ReturningPeer     bool
	Allowed           bool
	RequiresApproval  bool
	TemporarilyUnsafe bool
	TrustEntryState   string
}

// ReputationChecker returns a trust assessment for a peer DID.
type ReputationChecker func(ctx context.Context, peerDID string) (*ReputationAssessment, error)

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
		rules:    make([]ACLRule, 0, len(rules)),
		limiters: make(map[string]*rate.Limiter),
		logger:   logger,
	}

	// Initialize from provided rules; warn on overly permissive ones
	// but still load them for backward compatibility.
	for _, r := range rules {
		if err := ValidateRule(r); err != nil {
			logger.Warnw("loading overly permissive firewall rule (consider removing)",
				"peerDID", r.PeerDID,
				"action", r.Action,
				"tools", r.Tools,
				"warning", err.Error(),
			)
		}
		f.rules = append(f.rules, r)
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
			return fmt.Errorf("%w for peer %s", ErrRateLimitExceeded, peerDID)
		}
	}
	// Also check wildcard rate limiter.
	if limiter, ok := f.limiters[WildcardAll]; ok {
		if !limiter.Allow() {
			return ErrGlobalRateLimitExceeded
		}
	}

	// Check reputation score and trust-entry state.
	if f.reputationCheck != nil {
		assessment, err := f.reputationCheck(ctx, peerDID)
		if err != nil {
			f.logger.Warnw("reputation check error", "peerDID", peerDID, "error", err)
			// Don't block on reputation errors, continue to ACL.
		} else if assessment != nil {
			switch {
			case assessment.TemporarilyUnsafe:
				return fmt.Errorf("peer %s is temporarily unsafe for runtime admission", peerDID)
			case assessment.ReturningPeer && (!assessment.Allowed || assessment.RequiresApproval):
				return fmt.Errorf("peer %s trust entry %s requires review", peerDID, assessment.TrustEntryState)
			case assessment.Score > 0 && assessment.Score < f.minTrustScore:
				// score == 0 means bootstrap peer, allow through to ACL checks.
				return fmt.Errorf("peer %s reputation %.2f below minimum %.2f", peerDID, assessment.Score, f.minTrustScore)
			}
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
		case ACLActionAllow:
			allowed = true
		case ACLActionDeny:
			return fmt.Errorf("%w for peer %s, tool %s", ErrQueryDenied, peerDID, toolName)
		}
	}

	if !allowed {
		return fmt.Errorf("%w for peer %s, tool %s", ErrNoMatchingAllowRule, peerDID, toolName)
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

// ValidateRule checks whether an ACL rule is safe to add. It rejects
// overly permissive allow rules (wildcard peer + wildcard tools).
func ValidateRule(rule ACLRule) error {
	if rule.Action != ACLActionAllow {
		return nil // deny rules are always safe
	}

	isWildcardPeer := rule.PeerDID == WildcardAll
	isWildcardTools := len(rule.Tools) == 0
	for _, t := range rule.Tools {
		if t == WildcardAll {
			isWildcardTools = true
			break
		}
	}

	if isWildcardPeer && isWildcardTools {
		return fmt.Errorf("overly permissive rule: allow all peers with all tools is prohibited")
	}

	return nil
}

// AddRule validates and adds a new ACL rule. Returns an error if the rule
// is overly permissive (e.g. allow * with all tools).
func (f *Firewall) AddRule(rule ACLRule) error {
	if err := ValidateRule(rule); err != nil {
		return err
	}

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
	return nil
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
	if pattern == WildcardAll {
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
		if p == WildcardAll {
			return true
		}
		if strings.HasSuffix(p, WildcardAll) {
			if strings.HasPrefix(toolName, strings.TrimSuffix(p, WildcardAll)) {
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
