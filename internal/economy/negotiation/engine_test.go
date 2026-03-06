package negotiation

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/langoai/lango/internal/config"
)

func fixedNow() time.Time {
	return time.Date(2026, 3, 6, 12, 0, 0, 0, time.UTC)
}

func testEngine() *Engine {
	e := New(config.NegotiationConfig{
		Enabled:   true,
		MaxRounds: 3,
		Timeout:   10 * time.Minute,
	})
	e.nowFunc = fixedNow
	return e
}

func testTerms(price int64) Terms {
	return Terms{
		Price:    big.NewInt(price),
		Currency: "USDC",
		ToolName: "code-review",
	}
}

// propose is a test helper that creates a session with a known ID.
func propose(e *Engine, initiator, responder string, terms Terms) *NegotiationSession {
	ctx := context.Background()
	s, _ := e.Propose(ctx, initiator, responder, terms)
	return s
}

func TestEngine_Propose(t *testing.T) {
	e := testEngine()
	ctx := context.Background()

	session, err := e.Propose(ctx, "did:buyer", "did:seller", testTerms(5000))
	if err != nil {
		t.Fatalf("Propose() error: %v", err)
	}

	if session.Phase != PhaseProposed {
		t.Errorf("Phase = %q, want %q", session.Phase, PhaseProposed)
	}
	if session.Round != 1 {
		t.Errorf("Round = %d, want 1", session.Round)
	}
	if len(session.Proposals) != 1 {
		t.Errorf("Proposals len = %d, want 1", len(session.Proposals))
	}
	if session.MaxRounds != 3 {
		t.Errorf("MaxRounds = %d, want 3", session.MaxRounds)
	}
	if session.CurrentTerms.Price.Cmp(big.NewInt(5000)) != 0 {
		t.Errorf("Price = %s, want 5000", session.CurrentTerms.Price)
	}
	if session.ID == "" {
		t.Error("expected non-empty session ID")
	}
}

func TestEngine_Counter(t *testing.T) {
	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))

	session, err := e.Counter(ctx, s.ID, "did:seller", testTerms(4000), "too expensive")
	if err != nil {
		t.Fatalf("Counter() error: %v", err)
	}

	if session.Phase != PhaseCountered {
		t.Errorf("Phase = %q, want %q", session.Phase, PhaseCountered)
	}
	if session.Round != 2 {
		t.Errorf("Round = %d, want 2", session.Round)
	}
	if len(session.Proposals) != 2 {
		t.Errorf("Proposals len = %d, want 2", len(session.Proposals))
	}
	if session.CurrentTerms.Price.Cmp(big.NewInt(4000)) != 0 {
		t.Errorf("Price = %s, want 4000", session.CurrentTerms.Price)
	}
}

func TestEngine_Accept(t *testing.T) {
	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))

	session, err := e.Accept(ctx, s.ID, "did:seller")
	if err != nil {
		t.Fatalf("Accept() error: %v", err)
	}

	if session.Phase != PhaseAccepted {
		t.Errorf("Phase = %q, want %q", session.Phase, PhaseAccepted)
	}
	if !session.IsTerminal() {
		t.Error("expected session to be terminal after accept")
	}
}

func TestEngine_Reject(t *testing.T) {
	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))

	session, err := e.Reject(ctx, s.ID, "did:seller", "too expensive")
	if err != nil {
		t.Fatalf("Reject() error: %v", err)
	}

	if session.Phase != PhaseRejected {
		t.Errorf("Phase = %q, want %q", session.Phase, PhaseRejected)
	}
	if !session.IsTerminal() {
		t.Error("expected session to be terminal after reject")
	}
	last := session.Proposals[len(session.Proposals)-1]
	if last.Reason != "too expensive" {
		t.Errorf("Reason = %q, want %q", last.Reason, "too expensive")
	}
}

func TestEngine_Cancel(t *testing.T) {
	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))

	session, err := e.Cancel(ctx, s.ID, "did:buyer")
	if err != nil {
		t.Fatalf("Cancel() error: %v", err)
	}

	if session.Phase != PhaseCancelled {
		t.Errorf("Phase = %q, want %q", session.Phase, PhaseCancelled)
	}
}

func TestEngine_Cancel_OnlyInitiator(t *testing.T) {
	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))

	_, err := e.Cancel(ctx, s.ID, "did:seller")
	if !errors.Is(err, ErrInvalidSender) {
		t.Errorf("Cancel by responder: got %v, want ErrInvalidSender", err)
	}
}

func TestEngine_TurnEnforcement(t *testing.T) {
	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))

	// Buyer proposed, buyer cannot counter immediately
	_, err := e.Counter(ctx, s.ID, "did:buyer", testTerms(4500), "lower")
	if !errors.Is(err, ErrNotYourTurn) {
		t.Errorf("same-sender counter: got %v, want ErrNotYourTurn", err)
	}

	// Seller counters
	e.Counter(ctx, s.ID, "did:seller", testTerms(4000), "counter")

	// Seller cannot counter again
	_, err = e.Counter(ctx, s.ID, "did:seller", testTerms(3500), "again")
	if !errors.Is(err, ErrNotYourTurn) {
		t.Errorf("double counter: got %v, want ErrNotYourTurn", err)
	}
}

func TestEngine_MaxRounds(t *testing.T) {
	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))
	e.Counter(ctx, s.ID, "did:seller", testTerms(4000), "r2")
	e.Counter(ctx, s.ID, "did:buyer", testTerms(4500), "r3")

	// Round 3 == MaxRounds, no more counters
	_, err := e.Counter(ctx, s.ID, "did:seller", testTerms(4200), "r4")
	if !errors.Is(err, ErrMaxRoundsReached) {
		t.Errorf("beyond max rounds: got %v, want ErrMaxRoundsReached", err)
	}
}

func TestEngine_TerminalReject(t *testing.T) {
	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))
	e.Reject(ctx, s.ID, "did:seller", "no")

	_, err := e.Counter(ctx, s.ID, "did:buyer", testTerms(4000), "try again")
	if !errors.Is(err, ErrSessionTerminal) {
		t.Errorf("counter after reject: got %v, want ErrSessionTerminal", err)
	}
}

func TestEngine_SessionNotFound(t *testing.T) {
	e := testEngine()
	ctx := context.Background()

	_, err := e.Accept(ctx, "nonexistent", "did:x")
	if !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("got %v, want ErrSessionNotFound", err)
	}
}

func TestEngine_Expiry(t *testing.T) {
	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))

	// Advance time past expiry
	e.nowFunc = func() time.Time {
		return fixedNow().Add(15 * time.Minute)
	}

	_, err := e.Accept(ctx, s.ID, "did:seller")
	if !errors.Is(err, ErrSessionExpired) {
		t.Errorf("expired session: got %v, want ErrSessionExpired", err)
	}
}

func TestEngine_CheckExpiry(t *testing.T) {
	e := testEngine()
	ctx := context.Background()

	s1 := propose(e, "did:buyer", "did:seller", testTerms(5000))
	s2 := propose(e, "did:buyer2", "did:seller2", testTerms(3000))
	// Accept s2 so it won't be expired
	e.Accept(ctx, s2.ID, "did:seller2")

	e.nowFunc = func() time.Time {
		return fixedNow().Add(15 * time.Minute)
	}

	expired := e.CheckExpiry()
	if len(expired) != 1 {
		t.Errorf("CheckExpiry() len = %d, want 1", len(expired))
	}
	if len(expired) > 0 && expired[0] != s1.ID {
		t.Errorf("expired ID = %q, want %q", expired[0], s1.ID)
	}

	got, _ := e.Get(s1.ID)
	if got.Phase != PhaseExpired {
		t.Errorf("s1 Phase = %q, want %q", got.Phase, PhaseExpired)
	}

	got2, _ := e.Get(s2.ID)
	if got2.Phase != PhaseAccepted {
		t.Errorf("s2 Phase = %q, want %q", got2.Phase, PhaseAccepted)
	}
}

func TestEngine_Get_And_List(t *testing.T) {
	e := testEngine()

	s1 := propose(e, "did:a", "did:b", testTerms(1000))
	propose(e, "did:c", "did:d", testTerms(2000))

	s, err := e.Get(s1.ID)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if s.ID != s1.ID {
		t.Errorf("ID = %q, want %q", s.ID, s1.ID)
	}

	all := e.List()
	if len(all) != 2 {
		t.Errorf("List() len = %d, want 2", len(all))
	}
}

func TestEngine_ListByPeer(t *testing.T) {
	e := testEngine()

	propose(e, "did:alice", "did:bob", testTerms(1000))
	propose(e, "did:alice", "did:carol", testTerms(2000))
	propose(e, "did:dave", "did:eve", testTerms(3000))

	aliceSessions := e.ListByPeer("did:alice")
	if len(aliceSessions) != 2 {
		t.Errorf("ListByPeer(alice) len = %d, want 2", len(aliceSessions))
	}

	bobSessions := e.ListByPeer("did:bob")
	if len(bobSessions) != 1 {
		t.Errorf("ListByPeer(bob) len = %d, want 1", len(bobSessions))
	}

	nobodySessions := e.ListByPeer("did:nobody")
	if len(nobodySessions) != 0 {
		t.Errorf("ListByPeer(nobody) len = %d, want 0", len(nobodySessions))
	}
}

func TestEngine_FullNegotiationFlow(t *testing.T) {
	e := testEngine()
	ctx := context.Background()

	// Buyer proposes at 5000
	s, _ := e.Propose(ctx, "did:buyer", "did:seller", testTerms(5000))
	if s.Phase != PhaseProposed {
		t.Fatalf("Phase = %q, want proposed", s.Phase)
	}

	// Seller counters at 3000
	s, _ = e.Counter(ctx, s.ID, "did:seller", testTerms(3000), "lower please")
	if s.Phase != PhaseCountered {
		t.Fatalf("Phase = %q, want countered", s.Phase)
	}

	// Buyer counters at 4000
	s, _ = e.Counter(ctx, s.ID, "did:buyer", testTerms(4000), "meet in middle")
	if s.Phase != PhaseCountered {
		t.Fatalf("Phase = %q, want countered", s.Phase)
	}

	// Seller accepts
	s, _ = e.Accept(ctx, s.ID, "did:seller")
	if s.Phase != PhaseAccepted {
		t.Fatalf("Phase = %q, want accepted", s.Phase)
	}
	if s.CurrentTerms.Price.Cmp(big.NewInt(4000)) != 0 {
		t.Errorf("final price = %s, want 4000", s.CurrentTerms.Price)
	}
	if len(s.Proposals) != 4 {
		t.Errorf("proposal count = %d, want 4", len(s.Proposals))
	}
}

func TestEngine_ThirdPartyRejected(t *testing.T) {
	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))

	_, err := e.Accept(ctx, s.ID, "did:stranger")
	if !errors.Is(err, ErrInvalidSender) {
		t.Errorf("third-party accept: got %v, want ErrInvalidSender", err)
	}
}

func TestEngine_EventCallback(t *testing.T) {
	e := testEngine()
	ctx := context.Background()

	var events []Phase
	e.SetEventCallback(func(_ string, phase Phase) {
		events = append(events, phase)
	})

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))
	e.Accept(ctx, s.ID, "did:seller")

	if len(events) != 2 {
		t.Fatalf("events len = %d, want 2", len(events))
	}
	if events[0] != PhaseProposed {
		t.Errorf("events[0] = %q, want %q", events[0], PhaseProposed)
	}
	if events[1] != PhaseAccepted {
		t.Errorf("events[1] = %q, want %q", events[1], PhaseAccepted)
	}
}

func TestEngine_SetPricing(t *testing.T) {
	e := testEngine()
	called := false
	e.SetPricing(func(toolName string, peerDID string) (*big.Int, error) {
		called = true
		return big.NewInt(5000), nil
	})

	if e.pricing == nil {
		t.Error("expected pricing to be set")
	}
	_, _ = e.pricing("test", "did:x")
	if !called {
		t.Error("expected pricing function to be called")
	}
}

func TestEngine_AutoRespond_AcceptGoodPrice(t *testing.T) {
	e := New(config.NegotiationConfig{
		Enabled:       true,
		MaxRounds:     3,
		Timeout:       10 * time.Minute,
		AutoNegotiate: true,
		MaxDiscount:   0.2,
	})
	e.nowFunc = fixedNow
	e.SetPricing(func(_ string, _ string) (*big.Int, error) {
		return big.NewInt(5000), nil // base price 5000
	})
	ctx := context.Background()

	// Buyer proposes at 5000 (== base price), should auto-accept
	s := propose(e, "did:buyer", "did:seller", testTerms(5000))
	result, err := e.AutoRespond(ctx, s.ID)
	if err != nil {
		t.Fatalf("AutoRespond() error: %v", err)
	}
	if result.Phase != PhaseAccepted {
		t.Errorf("Phase = %q, want %q", result.Phase, PhaseAccepted)
	}
}

func TestEngine_AutoRespond_AcceptWithinDiscount(t *testing.T) {
	e := New(config.NegotiationConfig{
		Enabled:       true,
		MaxRounds:     3,
		Timeout:       10 * time.Minute,
		AutoNegotiate: true,
		MaxDiscount:   0.2, // min acceptable = 4000
	})
	e.nowFunc = fixedNow
	e.SetPricing(func(_ string, _ string) (*big.Int, error) {
		return big.NewInt(5000), nil
	})
	ctx := context.Background()

	// Buyer proposes at 4500 (within 20% discount), should auto-accept
	s := propose(e, "did:buyer", "did:seller", testTerms(4500))
	result, err := e.AutoRespond(ctx, s.ID)
	if err != nil {
		t.Fatalf("AutoRespond() error: %v", err)
	}
	if result.Phase != PhaseAccepted {
		t.Errorf("Phase = %q, want %q", result.Phase, PhaseAccepted)
	}
}

func TestEngine_AutoRespond_CounterWhenNegotiable(t *testing.T) {
	e := New(config.NegotiationConfig{
		Enabled:       true,
		MaxRounds:     3,
		Timeout:       10 * time.Minute,
		AutoNegotiate: true,
		MaxDiscount:   0.2, // min acceptable = 4000
	})
	e.nowFunc = fixedNow
	e.SetPricing(func(_ string, _ string) (*big.Int, error) {
		return big.NewInt(5000), nil
	})
	ctx := context.Background()

	// Buyer proposes at 3000 (below min 4000), should counter
	s := propose(e, "did:buyer", "did:seller", testTerms(3000))
	result, err := e.AutoRespond(ctx, s.ID)
	if err != nil {
		t.Fatalf("AutoRespond() error: %v", err)
	}
	if result.Phase != PhaseCountered {
		t.Errorf("Phase = %q, want %q", result.Phase, PhaseCountered)
	}
	// Counter should be midpoint of proposed(3000) and base(5000) = 4000
	if result.CurrentTerms.Price.Cmp(big.NewInt(4000)) != 0 {
		t.Errorf("counter price = %s, want 4000", result.CurrentTerms.Price)
	}
}

func TestEngine_AutoRespond_RejectTooLowNoRounds(t *testing.T) {
	e := New(config.NegotiationConfig{
		Enabled:       true,
		MaxRounds:     1,
		Timeout:       10 * time.Minute,
		AutoNegotiate: true,
		MaxDiscount:   0.2,
	})
	e.nowFunc = fixedNow
	e.SetPricing(func(_ string, _ string) (*big.Int, error) {
		return big.NewInt(5000), nil
	})
	ctx := context.Background()

	// Buyer proposes at 1000, MaxRounds=1 (already at round 1, can't counter)
	s := propose(e, "did:buyer", "did:seller", testTerms(1000))
	result, err := e.AutoRespond(ctx, s.ID)
	if err != nil {
		t.Fatalf("AutoRespond() error: %v", err)
	}
	if result.Phase != PhaseRejected {
		t.Errorf("Phase = %q, want %q", result.Phase, PhaseRejected)
	}
}

func TestEngine_AutoRespond_NoPricing(t *testing.T) {
	e := New(config.NegotiationConfig{
		Enabled:       true,
		MaxRounds:     3,
		Timeout:       10 * time.Minute,
		AutoNegotiate: true,
	})
	e.nowFunc = fixedNow
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))
	result, err := e.AutoRespond(ctx, s.ID)
	if err != nil {
		t.Fatalf("AutoRespond() error: %v", err)
	}
	if result.Phase != PhaseRejected {
		t.Errorf("Phase = %q, want %q (no pricing)", result.Phase, PhaseRejected)
	}
}

func TestEngine_DefaultConfig(t *testing.T) {
	e := New(config.NegotiationConfig{})
	if e.cfg.MaxRounds != 5 {
		t.Errorf("MaxRounds = %d, want 5", e.cfg.MaxRounds)
	}
	if e.cfg.Timeout != 5*time.Minute {
		t.Errorf("Timeout = %v, want 5m", e.cfg.Timeout)
	}
}
