package memory

import (
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/langoai/lango/internal/graph"
)

func TestGraphHooks_OnObservation(t *testing.T) {
	var received []graph.Triple
	hooks := NewGraphHooks(func(triples []graph.Triple) {
		received = append(received, triples...)
	}, zap.NewNop().Sugar())

	obs := Observation{
		ID:         uuid.New(),
		SessionKey: "session-1",
		Content:    "test observation",
	}

	t.Run("without previous", func(t *testing.T) {
		received = nil
		hooks.OnObservation(obs, "")

		if len(received) != 1 {
			t.Fatalf("want 1 triple, got %d", len(received))
		}
		if received[0].Predicate != graph.InSession {
			t.Errorf("want InSession, got %q", received[0].Predicate)
		}
	})

	t.Run("with previous", func(t *testing.T) {
		received = nil
		hooks.OnObservation(obs, "prev-id")

		if len(received) != 2 {
			t.Fatalf("want 2 triples, got %d", len(received))
		}

		var hasInSession, hasFollows bool
		for _, triple := range received {
			if triple.Predicate == graph.InSession {
				hasInSession = true
			}
			if triple.Predicate == graph.Follows {
				hasFollows = true
				if triple.Object != "observation:prev-id" {
					t.Errorf("want object 'observation:prev-id', got %q", triple.Object)
				}
			}
		}
		if !hasInSession {
			t.Error("want InSession triple")
		}
		if !hasFollows {
			t.Error("want Follows triple")
		}
	})
}

func TestGraphHooks_OnReflection(t *testing.T) {
	var received []graph.Triple
	hooks := NewGraphHooks(func(triples []graph.Triple) {
		received = append(received, triples...)
	}, zap.NewNop().Sugar())

	ref := Reflection{
		ID:         uuid.New(),
		SessionKey: "session-1",
		Content:    "test reflection",
	}

	t.Run("without observations", func(t *testing.T) {
		received = nil
		hooks.OnReflection(ref, nil)

		if len(received) != 1 {
			t.Fatalf("want 1 triple, got %d", len(received))
		}
		if received[0].Predicate != graph.InSession {
			t.Errorf("want InSession, got %q", received[0].Predicate)
		}
	})

	t.Run("with observations", func(t *testing.T) {
		received = nil
		hooks.OnReflection(ref, []string{"obs-1", "obs-2"})

		if len(received) != 3 {
			t.Fatalf("want 3 triples (1 InSession + 2 ReflectsOn), got %d", len(received))
		}

		reflectsOnCount := 0
		for _, triple := range received {
			if triple.Predicate == graph.ReflectsOn {
				reflectsOnCount++
			}
		}
		if reflectsOnCount != 2 {
			t.Errorf("want 2 ReflectsOn triples, got %d", reflectsOnCount)
		}
	})
}

func TestGraphHooks_NilCallback(t *testing.T) {
	hooks := NewGraphHooks(nil, zap.NewNop().Sugar())

	obs := Observation{
		ID:         uuid.New(),
		SessionKey: "session-1",
	}
	ref := Reflection{
		ID:         uuid.New(),
		SessionKey: "session-1",
	}

	// Should not panic with nil callback.
	hooks.OnObservation(obs, "")
	hooks.OnReflection(ref, []string{"obs-1"})
}
