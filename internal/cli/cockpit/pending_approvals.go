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
	queued  []chat.ApprovalRequestMsg
}

func NewPendingApprovalRegistry() *PendingApprovalRegistry {
	return &PendingApprovalRegistry{}
}

// Register records a pending approval request. Only one request remains
// visible at a time; later arrivals wait in FIFO order behind the current head.
func (r *PendingApprovalRegistry) Register(msg chat.ApprovalRequestMsg) {
	r.mu.Lock()
	msgCopy := msg
	msgCopy.Request.ToolName = sanitizeMissionProjectionText(msgCopy.Request.ToolName)
	msgCopy.Request.Summary = sanitizeMissionProjectionText(msgCopy.Request.Summary)
	msgCopy.ViewModel.RuleExplanation = sanitizeMissionProjectionText(msgCopy.ViewModel.RuleExplanation)
	msgCopy.ViewModel.Risk.Level = sanitizeMissionProjectionText(msgCopy.ViewModel.Risk.Level)
	msgCopy.ViewModel.Risk.Label = sanitizeMissionProjectionText(msgCopy.ViewModel.Risk.Label)
	if r.pending == nil {
		r.pending = &msgCopy
	} else {
		r.queued = append(r.queued, msgCopy)
	}
	r.mu.Unlock()
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
	if len(r.queued) > 0 {
		next := r.queued[0]
		r.queued[0] = chat.ApprovalRequestMsg{}
		r.queued = r.queued[1:]
		r.pending = &next
	} else {
		r.pending = nil
	}
	r.mu.Unlock()

	msg.Response <- resp
	return true
}
