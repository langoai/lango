package learning

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/eventbus"
)

func TestSuggestionEmitter_BelowThresholdSuppressed(t *testing.T) {
	t.Parallel()

	em := NewSuggestionEmitter(nil, 0.5, 1, time.Hour)
	em.TickTurn("s1")

	emitted := em.MaybeEmit(context.Background(), SuggestionCandidate{
		SessionKey: "s1",
		Pattern:    "p",
		Confidence: 0.4,
	})
	assert.False(t, emitted)
}

func TestSuggestionEmitter_RateLimitSuppressed(t *testing.T) {
	t.Parallel()

	em := NewSuggestionEmitter(nil, 0.5, 3, time.Hour)
	// One tick only; rate-limit requires >= 3.
	em.TickTurn("s1")

	emitted := em.MaybeEmit(context.Background(), SuggestionCandidate{
		SessionKey: "s1",
		Pattern:    "p",
		Confidence: 0.8,
	})
	assert.False(t, emitted)
}

func TestSuggestionEmitter_EmitsAboveThresholdAfterTicks(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	received := make(chan eventbus.LearningSuggestionEvent, 1)
	eventbus.SubscribeTyped(bus, func(e eventbus.LearningSuggestionEvent) {
		received <- e
	})

	em := NewSuggestionEmitter(bus, 0.5, 2, time.Hour)
	em.TickTurn("s1")
	em.TickTurn("s1")

	emitted := em.MaybeEmit(context.Background(), SuggestionCandidate{
		SessionKey: "s1",
		Pattern:    "p-unique",
		Confidence: 0.7,
	})
	assert.True(t, emitted)
	select {
	case e := <-received:
		assert.Equal(t, "s1", e.SessionKey)
		assert.Equal(t, "p-unique", e.Pattern)
	case <-time.After(time.Second):
		t.Fatal("event not received")
	}
}

func TestSuggestionEmitter_DedupWithinWindow(t *testing.T) {
	t.Parallel()

	em := NewSuggestionEmitter(nil, 0.5, 1, time.Hour)
	em.TickTurn("s1")

	ok1 := em.MaybeEmit(context.Background(), SuggestionCandidate{
		SessionKey: "s1",
		Pattern:    "same",
		Confidence: 0.7,
	})
	assert.True(t, ok1)

	// Reset turn counter to make rate-limit irrelevant.
	em.TickTurn("s1")

	ok2 := em.MaybeEmit(context.Background(), SuggestionCandidate{
		SessionKey: "s1",
		Pattern:    "same",
		Confidence: 0.7,
	})
	assert.False(t, ok2, "dedup should suppress the second emit within the window")
}

func TestSuggestionEmitter_DismissSuppressesReemission(t *testing.T) {
	t.Parallel()

	em := NewSuggestionEmitter(nil, 0.5, 1, time.Hour)
	em.TickTurn("s1")
	em.Dismiss("same-pattern")

	ok := em.MaybeEmit(context.Background(), SuggestionCandidate{
		SessionKey: "s1",
		Pattern:    "same-pattern",
		Confidence: 0.7,
	})
	assert.False(t, ok, "dismissed pattern should not re-emit within the dedup window")
}

func TestEmitSpecDrift_BelowThreshold(t *testing.T) {
	t.Parallel()

	em := NewSuggestionEmitter(nil, 0.5, 1, time.Hour)

	for i := 0; i < defaultDriftThreshold-1; i++ {
		ok := em.EmitSpecDrift(context.Background(), "exec", "timeout", "context deadline exceeded")
		assert.False(t, ok, "should not emit below threshold")
	}
}

func TestEmitSpecDrift_ThresholdCrossed(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	received := make(chan eventbus.SpecDriftDetectedEvent, 1)
	eventbus.SubscribeTyped(bus, func(e eventbus.SpecDriftDetectedEvent) {
		received <- e
	})

	em := NewSuggestionEmitter(bus, 0.5, 1, time.Hour)

	var emitted bool
	for i := 0; i < defaultDriftThreshold; i++ {
		emitted = em.EmitSpecDrift(context.Background(), "exec", "timeout", "deadline exceeded")
	}
	assert.True(t, emitted, "should emit at threshold")

	select {
	case e := <-received:
		assert.Equal(t, "exec", e.ToolName)
		assert.Equal(t, "timeout", e.ErrorClass)
		assert.Equal(t, defaultDriftThreshold, e.Occurrences)
	case <-time.After(time.Second):
		t.Fatal("event not received")
	}
}

func TestEmitSpecDrift_DedupAfterEmit(t *testing.T) {
	t.Parallel()

	em := NewSuggestionEmitter(nil, 0.5, 1, time.Hour)

	for i := 0; i < defaultDriftThreshold; i++ {
		em.EmitSpecDrift(context.Background(), "exec", "timeout", "err")
	}

	ok := em.EmitSpecDrift(context.Background(), "exec", "timeout", "err")
	assert.False(t, ok, "dedup should suppress immediate re-emission")
}
