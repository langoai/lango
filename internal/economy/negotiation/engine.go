package negotiation

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/langoai/lango/internal/config"
)

var (
	ErrSessionNotFound  = errors.New("negotiation session not found")
	ErrSessionTerminal  = errors.New("negotiation session is in terminal state")
	ErrNotYourTurn      = errors.New("not your turn to act")
	ErrMaxRoundsReached = errors.New("maximum negotiation rounds reached")
	ErrSessionExpired   = errors.New("session expired")
	ErrInvalidSender    = errors.New("sender is not a participant")
)

// PricingQuerier queries tool pricing. Defined locally to avoid import cycles.
type PricingQuerier func(toolName string, peerDID string) (*big.Int, error)

// EventCallback is called on negotiation state changes.
type EventCallback func(sessionID string, phase Phase)

// Engine manages negotiation sessions.
type Engine struct {
	mu       sync.RWMutex
	sessions map[string]*NegotiationSession
	cfg      config.NegotiationConfig
	pricing  PricingQuerier
	onEvent  EventCallback
	nowFunc  func() time.Time
}

// New creates a new negotiation engine.
func New(cfg config.NegotiationConfig) *Engine {
	if cfg.MaxRounds <= 0 {
		cfg.MaxRounds = 5
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 5 * time.Minute
	}
	return &Engine{
		sessions: make(map[string]*NegotiationSession),
		cfg:      cfg,
		nowFunc:  time.Now,
	}
}

// SetPricing sets the pricing query function.
func (e *Engine) SetPricing(fn PricingQuerier) {
	e.pricing = fn
}

// SetEventCallback sets the callback for negotiation state changes.
func (e *Engine) SetEventCallback(fn EventCallback) {
	e.onEvent = fn
}

// Propose starts a new negotiation session.
func (e *Engine) Propose(ctx context.Context, initiatorDID, responderDID string, terms Terms) (*NegotiationSession, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := e.nowFunc()
	sessionID := uuid.New().String()

	proposal := Proposal{
		Action:    ActionPropose,
		SenderDID: initiatorDID,
		Terms:     terms,
		Round:     1,
		Timestamp: now,
	}

	session := &NegotiationSession{
		ID:           sessionID,
		InitiatorDID: initiatorDID,
		ResponderDID: responderDID,
		Phase:        PhaseProposed,
		CurrentTerms: &terms,
		Proposals:    []Proposal{proposal},
		Round:        1,
		MaxRounds:    e.cfg.MaxRounds,
		CreatedAt:    now,
		UpdatedAt:    now,
		ExpiresAt:    now.Add(e.cfg.Timeout),
	}

	e.sessions[sessionID] = session
	e.fireEvent(sessionID, PhaseProposed)
	return session, nil
}

// Counter submits a counter-offer to an existing session.
func (e *Engine) Counter(ctx context.Context, sessionID, senderDID string, terms Terms, reason string) (*NegotiationSession, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	session, err := e.getAndValidate(sessionID, senderDID)
	if err != nil {
		return nil, fmt.Errorf("counter %q: %w", sessionID, err)
	}

	if !session.CanCounter() {
		return nil, fmt.Errorf("counter %q: %w", sessionID, ErrMaxRoundsReached)
	}

	now := e.nowFunc()
	session.Round++
	proposal := Proposal{
		Action:    ActionCounter,
		SenderDID: senderDID,
		Terms:     terms,
		Round:     session.Round,
		Reason:    reason,
		Timestamp: now,
	}

	session.Phase = PhaseCountered
	session.CurrentTerms = &terms
	session.Proposals = append(session.Proposals, proposal)
	session.UpdatedAt = now

	e.fireEvent(sessionID, PhaseCountered)
	return session, nil
}

// Accept accepts the current terms.
func (e *Engine) Accept(ctx context.Context, sessionID, senderDID string) (*NegotiationSession, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	session, err := e.getAndValidate(sessionID, senderDID)
	if err != nil {
		return nil, fmt.Errorf("accept %q: %w", sessionID, err)
	}

	now := e.nowFunc()
	proposal := Proposal{
		Action:    ActionAccept,
		SenderDID: senderDID,
		Terms:     *session.CurrentTerms,
		Round:     session.Round,
		Timestamp: now,
	}

	session.Phase = PhaseAccepted
	session.Proposals = append(session.Proposals, proposal)
	session.UpdatedAt = now

	e.fireEvent(sessionID, PhaseAccepted)
	return session, nil
}

// Reject rejects and terminates the negotiation.
func (e *Engine) Reject(ctx context.Context, sessionID, senderDID string, reason string) (*NegotiationSession, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	session, err := e.getAndValidate(sessionID, senderDID)
	if err != nil {
		return nil, fmt.Errorf("reject %q: %w", sessionID, err)
	}

	now := e.nowFunc()
	proposal := Proposal{
		Action:    ActionReject,
		SenderDID: senderDID,
		Terms:     *session.CurrentTerms,
		Round:     session.Round,
		Reason:    reason,
		Timestamp: now,
	}

	session.Phase = PhaseRejected
	session.Proposals = append(session.Proposals, proposal)
	session.UpdatedAt = now

	e.fireEvent(sessionID, PhaseRejected)
	return session, nil
}

// Cancel terminates a session by its initiator.
func (e *Engine) Cancel(ctx context.Context, sessionID, senderDID string) (*NegotiationSession, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	session, ok := e.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("cancel %q: %w", sessionID, ErrSessionNotFound)
	}
	if session.IsTerminal() {
		return nil, fmt.Errorf("cancel %q: %w", sessionID, ErrSessionTerminal)
	}
	if session.InitiatorDID != senderDID {
		return nil, fmt.Errorf("cancel %q: %w", sessionID, ErrInvalidSender)
	}

	session.Phase = PhaseCancelled
	session.UpdatedAt = e.nowFunc()

	e.fireEvent(sessionID, PhaseCancelled)
	return session, nil
}

// Get returns a session by ID.
func (e *Engine) Get(sessionID string) (*NegotiationSession, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	session, ok := e.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("get %q: %w", sessionID, ErrSessionNotFound)
	}
	return session, nil
}

// List returns all sessions.
func (e *Engine) List() []*NegotiationSession {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]*NegotiationSession, 0, len(e.sessions))
	for _, s := range e.sessions {
		result = append(result, s)
	}
	return result
}

// ListByPeer returns all sessions involving a peer.
func (e *Engine) ListByPeer(peerDID string) []*NegotiationSession {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var result []*NegotiationSession
	for _, s := range e.sessions {
		if s.InitiatorDID == peerDID || s.ResponderDID == peerDID {
			result = append(result, s)
		}
	}
	return result
}

// CheckExpiry expires timed-out sessions and returns their IDs.
func (e *Engine) CheckExpiry() []string {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := e.nowFunc()
	var expired []string
	for _, s := range e.sessions {
		if !s.IsTerminal() && now.After(s.ExpiresAt) {
			s.Phase = PhaseExpired
			s.UpdatedAt = now
			expired = append(expired, s.ID)
		}
	}

	for _, id := range expired {
		e.fireEventLocked(id, PhaseExpired)
	}

	return expired
}

// AutoRespond handles automatic negotiation responses for a session.
// It uses the configured maxDiscount and pricing to decide: accept, counter, or reject.
func (e *Engine) AutoRespond(ctx context.Context, sessionID string) (*NegotiationSession, error) {
	e.mu.RLock()
	session, ok := e.sessions[sessionID]
	if !ok {
		e.mu.RUnlock()
		return nil, fmt.Errorf("auto-respond %q: %w", sessionID, ErrSessionNotFound)
	}
	if session.IsTerminal() {
		e.mu.RUnlock()
		return nil, fmt.Errorf("auto-respond %q: %w", sessionID, ErrSessionTerminal)
	}

	// Determine our DID (we are the responder).
	responderDID := session.ResponderDID
	lastProposal := session.Proposals[len(session.Proposals)-1]

	// If it's not our turn, nothing to do.
	if lastProposal.SenderDID == responderDID {
		e.mu.RUnlock()
		return nil, fmt.Errorf("auto-respond %q: %w", sessionID, ErrNotYourTurn)
	}

	// Get base price from pricing function.
	var basePrice *big.Int
	if e.pricing != nil {
		p, err := e.pricing(lastProposal.Terms.ToolName, session.InitiatorDID)
		if err == nil {
			basePrice = p
		}
	}
	e.mu.RUnlock()

	if basePrice == nil {
		return e.Reject(ctx, sessionID, responderDID, "no base price available")
	}

	proposedPrice := lastProposal.Terms.Price

	// Accept if proposed >= basePrice.
	if proposedPrice.Cmp(basePrice) >= 0 {
		return e.Accept(ctx, sessionID, responderDID)
	}

	// Compute minimum acceptable price: (1 - maxDiscount) * basePrice.
	maxDiscount := e.cfg.MaxDiscount
	if maxDiscount <= 0 {
		maxDiscount = 0.2
	}
	// floorBps = (1.0 - maxDiscount) * 10000
	floorBps := int64((1.0 - maxDiscount) * 10000)
	minPrice := new(big.Int).Mul(basePrice, big.NewInt(floorBps))
	minPrice.Div(minPrice, big.NewInt(10000))

	// Accept if proposed >= minPrice.
	if proposedPrice.Cmp(minPrice) >= 0 {
		return e.Accept(ctx, sessionID, responderDID)
	}

	// Counter if rounds remaining.
	e.mu.RLock()
	canCounter := session.CanCounter()
	e.mu.RUnlock()

	if canCounter {
		strategy := NewAutoStrategy(basePrice, maxDiscount)
		counterPrice := strategy.GenerateCounter(proposedPrice, session.Round, session.MaxRounds)
		counterTerms := lastProposal.Terms
		counterTerms.Price = counterPrice
		return e.Counter(ctx, sessionID, responderDID, counterTerms, "auto-counter")
	}

	return e.Reject(ctx, sessionID, responderDID, "price too low, no rounds remaining")
}

// getAndValidate returns the session and validates it for action.
// Caller must hold e.mu.
func (e *Engine) getAndValidate(sessionID, senderDID string) (*NegotiationSession, error) {
	session, ok := e.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}
	if session.IsTerminal() {
		return nil, ErrSessionTerminal
	}
	if e.nowFunc().After(session.ExpiresAt) {
		session.Phase = PhaseExpired
		session.UpdatedAt = e.nowFunc()
		e.fireEventLocked(sessionID, PhaseExpired)
		return nil, ErrSessionExpired
	}
	if !isParticipant(session, senderDID) {
		return nil, ErrInvalidSender
	}
	if !isValidTurn(session, senderDID) {
		return nil, ErrNotYourTurn
	}
	return session, nil
}

// isParticipant checks that the sender is one of the participants.
func isParticipant(session *NegotiationSession, senderDID string) bool {
	return senderDID == session.InitiatorDID || senderDID == session.ResponderDID
}

// isValidTurn checks that the last proposal sender is not the same as the current sender.
func isValidTurn(session *NegotiationSession, senderDID string) bool {
	if len(session.Proposals) == 0 {
		return true
	}
	last := session.Proposals[len(session.Proposals)-1]
	return last.SenderDID != senderDID
}

// fireEvent calls the event callback if set. Caller must NOT hold e.mu write lock.
func (e *Engine) fireEvent(sessionID string, phase Phase) {
	if e.onEvent != nil {
		e.onEvent(sessionID, phase)
	}
}

// fireEventLocked calls the event callback if set. Safe to call while holding e.mu.
func (e *Engine) fireEventLocked(sessionID string, phase Phase) {
	if e.onEvent != nil {
		e.onEvent(sessionID, phase)
	}
}
