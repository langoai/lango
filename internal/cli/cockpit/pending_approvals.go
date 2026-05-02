package cockpit

import (
	"sync"

	"github.com/langoai/lango/internal/approval"
	"github.com/langoai/lango/internal/cli/chat"
)

// PendingApprovalRegistry stores the latest pending approval request for the
// lifetime of a cockpit session.
type PendingApprovalRegistry struct {
	mu      sync.RWMutex
	pending *chat.ApprovalRequestMsg
}

func NewPendingApprovalRegistry() *PendingApprovalRegistry {
	return &PendingApprovalRegistry{}
}

// Register replaces the latest pending approval request.
func (r *PendingApprovalRegistry) Register(msg chat.ApprovalRequestMsg) {
	var superseded *chat.ApprovalRequestMsg

	r.mu.Lock()
	if r.pending != nil {
		msgCopy := *r.pending
		superseded = &msgCopy
	}
	msgCopy := msg
	r.pending = &msgCopy
	r.mu.Unlock()

	if superseded != nil {
		superseded.Response <- approval.ApprovalResponse{
			Approved:    false,
			AlwaysAllow: false,
			Provider:    "tui",
		}
	}
}

// Latest returns the latest pending approval request, if any.
func (r *PendingApprovalRegistry) Latest() *chat.ApprovalRequestMsg {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.pending == nil {
		return nil
	}
	msgCopy := *r.pending
	return &msgCopy
}

// HasPending reports whether a pending approval exists.
func (r *PendingApprovalRegistry) HasPending() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.pending != nil
}

// Resolve writes a response to the original pending response channel exactly
// once when the request ID matches the latest pending request.
func (r *PendingApprovalRegistry) Resolve(id string, resp approval.ApprovalResponse) bool {
	r.mu.Lock()
	if r.pending == nil || r.pending.Request.ID != id {
		r.mu.Unlock()
		return false
	}
	msg := *r.pending
	r.pending = nil
	r.mu.Unlock()

	msg.Response <- resp
	return true
}
