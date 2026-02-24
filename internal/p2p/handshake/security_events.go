package handshake

import (
	"sync"

	"go.uber.org/zap"
)

// SecurityEventHandler tracks tool execution failures and reputation changes
// to auto-invalidate sessions when thresholds are exceeded.
type SecurityEventHandler struct {
	sessions      *SessionStore
	mu            sync.Mutex
	failureCounts map[string]int
	maxFailures   int
	minTrustScore float64
	logger        *zap.SugaredLogger
}

// NewSecurityEventHandler creates a handler that auto-invalidates sessions
// after consecutive tool failures or reputation drops below the threshold.
func NewSecurityEventHandler(
	sessions *SessionStore,
	maxFailures int,
	minTrustScore float64,
	logger *zap.SugaredLogger,
) *SecurityEventHandler {
	if maxFailures <= 0 {
		maxFailures = 5
	}
	return &SecurityEventHandler{
		sessions:      sessions,
		failureCounts: make(map[string]int),
		maxFailures:   maxFailures,
		minTrustScore: minTrustScore,
		logger:        logger,
	}
}

// RecordToolFailure increments the consecutive failure counter for the peer.
// When the counter reaches maxFailures, the session is auto-invalidated.
func (h *SecurityEventHandler) RecordToolFailure(peerDID string) {
	h.mu.Lock()
	h.failureCounts[peerDID]++
	count := h.failureCounts[peerDID]
	h.mu.Unlock()

	if count >= h.maxFailures {
		h.logger.Warnw("auto-invalidating session: repeated failures",
			"peerDID", peerDID, "failures", count)
		h.sessions.Invalidate(peerDID, ReasonRepeatedFailures)

		h.mu.Lock()
		delete(h.failureCounts, peerDID)
		h.mu.Unlock()
	}
}

// RecordToolSuccess resets the consecutive failure counter for the peer.
func (h *SecurityEventHandler) RecordToolSuccess(peerDID string) {
	h.mu.Lock()
	delete(h.failureCounts, peerDID)
	h.mu.Unlock()
}

// OnReputationChange invalidates the peer's session if the new score drops
// below the minimum trust threshold.
func (h *SecurityEventHandler) OnReputationChange(peerDID string, newScore float64) {
	if newScore < h.minTrustScore {
		h.logger.Warnw("auto-invalidating session: reputation drop",
			"peerDID", peerDID, "score", newScore, "threshold", h.minTrustScore)
		h.sessions.Invalidate(peerDID, ReasonReputationDrop)
	}
}
