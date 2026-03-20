package runledger

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/langoai/lango/internal/ent/enttest"
)

func BenchmarkEntStore_ParallelRuns(b *testing.B) {
	dsn := fmt.Sprintf("file:%s?_fk=1", filepath.Join(b.TempDir(), "ent-bench.db"))
	client := enttest.Open(b, "sqlite3", dsn)
	defer client.Close()

	store := NewEntStore(client)
	ctx := context.Background()
	runIDs := seedEntStoreBenchRuns(b, store, ctx, 8)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			runID := runIDs[i%len(runIDs)]
			if _, err := store.GetRunSnapshot(ctx, runID); err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

func BenchmarkEntStore_GlobalLock_Baseline(b *testing.B) {
	cache := newGlobalLockSnapshotCache()
	runIDs := make([]string, 8)
	for i := range runIDs {
		runIDs[i] = fmt.Sprintf("run-%d", i)
		cache.put(runIDs[i], (&RunSnapshot{RunID: runIDs[i]}).DeepCopy())
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = cache.get(runIDs[i%len(runIDs)])
			i++
		}
	})
}

func seedEntStoreBenchRuns(
	b *testing.B,
	store *EntStore,
	ctx context.Context,
	count int,
) []string {
	runIDs := make([]string, 0, count)
	for i := 0; i < count; i++ {
		runID := fmt.Sprintf("run-%d", i)
		runIDs = append(runIDs, runID)
		if err := store.AppendJournalEvent(ctx, JournalEvent{
			RunID:   runID,
			Type:    EventRunCreated,
			Payload: marshalPayload(RunCreatedPayload{SessionKey: "bench", Goal: runID}),
		}); err != nil {
			b.Fatal(err)
		}
		if err := store.AppendJournalEvent(ctx, JournalEvent{
			RunID: runID,
			Type:  EventPlanAttached,
			Payload: marshalPayload(PlanAttachedPayload{
				Steps: []Step{{
					StepID:     "s1",
					Goal:       "work",
					OwnerAgent: "operator",
					Status:     StepStatusPending,
					Validator:  ValidatorSpec{Type: ValidatorBuildPass},
					MaxRetries: DefaultMaxRetries,
				}},
			}),
		}); err != nil {
			b.Fatal(err)
		}
		if _, err := store.GetRunSnapshot(ctx, runID); err != nil {
			b.Fatal(err)
		}
	}
	return runIDs
}

type globalLockSnapshotCache struct {
	mu    sync.Mutex
	cache map[string]*RunSnapshot
}

func newGlobalLockSnapshotCache() *globalLockSnapshotCache {
	return &globalLockSnapshotCache{
		cache: make(map[string]*RunSnapshot),
	}
}

func (c *globalLockSnapshotCache) get(runID string) *RunSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cache[runID]
}

func (c *globalLockSnapshotCache) put(runID string, snap *RunSnapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[runID] = snap
}
