package negotiation

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	t.Parallel()

	e := testEngine()
	ctx := context.Background()

	session, err := e.Propose(ctx, "did:buyer", "did:seller", testTerms(5000))
	require.NoError(t, err)

	assert.Equal(t, PhaseProposed, session.Phase)
	assert.Equal(t, 1, session.Round)
	assert.Len(t, session.Proposals, 1)
	assert.Equal(t, 3, session.MaxRounds)
	assert.Equal(t, 0, session.CurrentTerms.Price.Cmp(big.NewInt(5000)))
	assert.NotEmpty(t, session.ID)
}

func TestEngine_Counter(t *testing.T) {
	t.Parallel()

	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))

	session, err := e.Counter(ctx, s.ID, "did:seller", testTerms(4000), "too expensive")
	require.NoError(t, err)

	assert.Equal(t, PhaseCountered, session.Phase)
	assert.Equal(t, 2, session.Round)
	assert.Len(t, session.Proposals, 2)
	assert.Equal(t, 0, session.CurrentTerms.Price.Cmp(big.NewInt(4000)))
}

func TestEngine_Accept(t *testing.T) {
	t.Parallel()

	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))

	session, err := e.Accept(ctx, s.ID, "did:seller")
	require.NoError(t, err)

	assert.Equal(t, PhaseAccepted, session.Phase)
	assert.True(t, session.IsTerminal())
}

func TestEngine_Reject(t *testing.T) {
	t.Parallel()

	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))

	session, err := e.Reject(ctx, s.ID, "did:seller", "too expensive")
	require.NoError(t, err)

	assert.Equal(t, PhaseRejected, session.Phase)
	assert.True(t, session.IsTerminal())
	last := session.Proposals[len(session.Proposals)-1]
	assert.Equal(t, "too expensive", last.Reason)
}

func TestEngine_Cancel(t *testing.T) {
	t.Parallel()

	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))

	session, err := e.Cancel(ctx, s.ID, "did:buyer")
	require.NoError(t, err)

	assert.Equal(t, PhaseCancelled, session.Phase)
}

func TestEngine_Cancel_OnlyInitiator(t *testing.T) {
	t.Parallel()

	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))

	_, err := e.Cancel(ctx, s.ID, "did:seller")
	require.ErrorIs(t, err, ErrInvalidSender)
}

func TestEngine_TurnEnforcement(t *testing.T) {
	t.Parallel()

	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))

	// Buyer proposed, buyer cannot counter immediately
	_, err := e.Counter(ctx, s.ID, "did:buyer", testTerms(4500), "lower")
	require.ErrorIs(t, err, ErrNotYourTurn)

	// Seller counters
	e.Counter(ctx, s.ID, "did:seller", testTerms(4000), "counter")

	// Seller cannot counter again
	_, err = e.Counter(ctx, s.ID, "did:seller", testTerms(3500), "again")
	require.ErrorIs(t, err, ErrNotYourTurn)
}

func TestEngine_MaxRounds(t *testing.T) {
	t.Parallel()

	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))
	e.Counter(ctx, s.ID, "did:seller", testTerms(4000), "r2")
	e.Counter(ctx, s.ID, "did:buyer", testTerms(4500), "r3")

	// Round 3 == MaxRounds, no more counters
	_, err := e.Counter(ctx, s.ID, "did:seller", testTerms(4200), "r4")
	require.ErrorIs(t, err, ErrMaxRoundsReached)
}

func TestEngine_TerminalReject(t *testing.T) {
	t.Parallel()

	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))
	e.Reject(ctx, s.ID, "did:seller", "no")

	_, err := e.Counter(ctx, s.ID, "did:buyer", testTerms(4000), "try again")
	require.ErrorIs(t, err, ErrSessionTerminal)
}

func TestEngine_SessionNotFound(t *testing.T) {
	t.Parallel()

	e := testEngine()
	ctx := context.Background()

	_, err := e.Accept(ctx, "nonexistent", "did:x")
	require.ErrorIs(t, err, ErrSessionNotFound)
}

func TestEngine_Expiry(t *testing.T) {
	t.Parallel()

	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))

	// Advance time past expiry
	e.nowFunc = func() time.Time {
		return fixedNow().Add(15 * time.Minute)
	}

	_, err := e.Accept(ctx, s.ID, "did:seller")
	require.ErrorIs(t, err, ErrSessionExpired)
}

func TestEngine_CheckExpiry(t *testing.T) {
	t.Parallel()

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
	require.Len(t, expired, 1)
	assert.Equal(t, s1.ID, expired[0])

	got, _ := e.Get(s1.ID)
	assert.Equal(t, PhaseExpired, got.Phase)

	got2, _ := e.Get(s2.ID)
	assert.Equal(t, PhaseAccepted, got2.Phase)
}

func TestEngine_Get_And_List(t *testing.T) {
	t.Parallel()

	e := testEngine()

	s1 := propose(e, "did:a", "did:b", testTerms(1000))
	propose(e, "did:c", "did:d", testTerms(2000))

	s, err := e.Get(s1.ID)
	require.NoError(t, err)
	assert.Equal(t, s1.ID, s.ID)

	all := e.List()
	assert.Len(t, all, 2)
}

func TestEngine_ListByPeer(t *testing.T) {
	t.Parallel()

	e := testEngine()

	propose(e, "did:alice", "did:bob", testTerms(1000))
	propose(e, "did:alice", "did:carol", testTerms(2000))
	propose(e, "did:dave", "did:eve", testTerms(3000))

	assert.Len(t, e.ListByPeer("did:alice"), 2)
	assert.Len(t, e.ListByPeer("did:bob"), 1)
	assert.Len(t, e.ListByPeer("did:nobody"), 0)
}

func TestEngine_FullNegotiationFlow(t *testing.T) {
	t.Parallel()

	e := testEngine()
	ctx := context.Background()

	// Buyer proposes at 5000
	s, _ := e.Propose(ctx, "did:buyer", "did:seller", testTerms(5000))
	require.Equal(t, PhaseProposed, s.Phase)

	// Seller counters at 3000
	s, _ = e.Counter(ctx, s.ID, "did:seller", testTerms(3000), "lower please")
	require.Equal(t, PhaseCountered, s.Phase)

	// Buyer counters at 4000
	s, _ = e.Counter(ctx, s.ID, "did:buyer", testTerms(4000), "meet in middle")
	require.Equal(t, PhaseCountered, s.Phase)

	// Seller accepts
	s, _ = e.Accept(ctx, s.ID, "did:seller")
	require.Equal(t, PhaseAccepted, s.Phase)
	assert.Equal(t, 0, s.CurrentTerms.Price.Cmp(big.NewInt(4000)))
	assert.Len(t, s.Proposals, 4)
}

func TestEngine_ThirdPartyRejected(t *testing.T) {
	t.Parallel()

	e := testEngine()
	ctx := context.Background()

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))

	_, err := e.Accept(ctx, s.ID, "did:stranger")
	require.ErrorIs(t, err, ErrInvalidSender)
}

func TestEngine_EventCallback(t *testing.T) {
	t.Parallel()

	e := testEngine()
	ctx := context.Background()

	var events []Phase
	e.SetEventCallback(func(_ string, phase Phase) {
		events = append(events, phase)
	})

	s := propose(e, "did:buyer", "did:seller", testTerms(5000))
	e.Accept(ctx, s.ID, "did:seller")

	require.Len(t, events, 2)
	assert.Equal(t, PhaseProposed, events[0])
	assert.Equal(t, PhaseAccepted, events[1])
}

func TestEngine_SetPricing(t *testing.T) {
	t.Parallel()

	e := testEngine()
	called := false
	e.SetPricing(func(toolName string, peerDID string) (*big.Int, error) {
		called = true
		return big.NewInt(5000), nil
	})

	require.NotNil(t, e.pricing)
	_, _ = e.pricing("test", "did:x")
	assert.True(t, called)
}

func TestEngine_AutoRespond_AcceptGoodPrice(t *testing.T) {
	t.Parallel()

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
	require.NoError(t, err)
	assert.Equal(t, PhaseAccepted, result.Phase)
}

func TestEngine_AutoRespond_AcceptWithinDiscount(t *testing.T) {
	t.Parallel()

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
	require.NoError(t, err)
	assert.Equal(t, PhaseAccepted, result.Phase)
}

func TestEngine_AutoRespond_CounterWhenNegotiable(t *testing.T) {
	t.Parallel()

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
	require.NoError(t, err)
	assert.Equal(t, PhaseCountered, result.Phase)
	// Counter should be midpoint of proposed(3000) and base(5000) = 4000
	assert.Equal(t, 0, result.CurrentTerms.Price.Cmp(big.NewInt(4000)))
}

func TestEngine_AutoRespond_RejectTooLowNoRounds(t *testing.T) {
	t.Parallel()

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
	require.NoError(t, err)
	assert.Equal(t, PhaseRejected, result.Phase)
}

func TestEngine_AutoRespond_NoPricing(t *testing.T) {
	t.Parallel()

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
	require.NoError(t, err)
	assert.Equal(t, PhaseRejected, result.Phase)
}

func TestEngine_DefaultConfig(t *testing.T) {
	t.Parallel()

	e := New(config.NegotiationConfig{})
	assert.Equal(t, 5, e.cfg.MaxRounds)
	assert.Equal(t, 5*time.Minute, e.cfg.Timeout)
}
