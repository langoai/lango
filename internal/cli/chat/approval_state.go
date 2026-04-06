package chat

import "time"

// approvalState encapsulates all approval-related state that was previously
// scattered across ChatModel fields and package-level globals.
type approvalState struct {
	pending        *ApprovalRequestMsg
	confirmPending bool
	confirmAction  string    // "a" or "s"
	confirmTime    time.Time
	scrollOffset   int  // was dialogScrollOffset (package global)
	splitMode      bool // was dialogSplitMode (package global)
}

// Reset prepares the approval state for a new approval request.
func (a *approvalState) Reset(msg *ApprovalRequestMsg) {
	a.pending = msg
	a.confirmPending = false
	a.confirmAction = ""
	a.scrollOffset = 0
	a.splitMode = false
}

// Clear removes the pending approval and resets confirmation state.
func (a *approvalState) Clear() {
	a.pending = nil
	a.confirmPending = false
	a.confirmAction = ""
}

// HasPending reports whether an approval request is currently pending.
func (a *approvalState) HasPending() bool {
	return a.pending != nil
}

// IsConfirmExpired reports whether the confirm-pending state has expired (>3s).
func (a *approvalState) IsConfirmExpired() bool {
	return a.confirmPending && time.Since(a.confirmTime) > 3*time.Second
}

// StartConfirm enters the confirm-pending state for the given action.
func (a *approvalState) StartConfirm(action string) {
	a.confirmPending = true
	a.confirmAction = action
	a.confirmTime = time.Now()
}

// CancelConfirm resets the confirm-pending state without clearing the approval.
func (a *approvalState) CancelConfirm() {
	a.confirmPending = false
	a.confirmAction = ""
}

// ScrollDiff adjusts the diff viewport scroll position, clamping to zero.
func (a *approvalState) ScrollDiff(delta int) {
	a.scrollOffset += delta
	if a.scrollOffset < 0 {
		a.scrollOffset = 0
	}
}

// ToggleSplit toggles between unified and split diff display modes.
func (a *approvalState) ToggleSplit() {
	a.splitMode = !a.splitMode
}
