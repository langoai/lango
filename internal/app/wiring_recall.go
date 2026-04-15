package app

import (
	"context"
	"sync"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/lifecycle"
	"github.com/langoai/lango/internal/session"
)

// wireSessionRecall constructs the RecallIndex, registers its startup sweep,
// and installs it as the session-end processor for hard-end flows.
func wireSessionRecall(app *App) *session.RecallIndex {
	if app == nil || app.Store == nil {
		return nil
	}
	cfg := app.Config.Context.Recall.ResolveRecall()
	if cfg.Enabled == nil || !*cfg.Enabled {
		return nil
	}
	entStore, ok := app.Store.(*session.EntStore)
	if !ok {
		return nil
	}
	// When the store was constructed via NewEntStoreWithClient (e.g., under
	// tests using testutil.TestEntClient), the raw sql.DB handle is not
	// available and FTS5 cannot be wired. Recall is silently disabled.
	if entStore.DB() == nil {
		return nil
	}
	idx := session.NewRecallIndex(entStore)
	// Eager EnsureReady to surface FTS5 availability early; if unavailable
	// (no-fts5 build), the retriever returns an error on query and we log
	// once here rather than on every turn.
	if err := idx.EnsureReady(); err != nil {
		logger().Warnw("session recall disabled — FTS5 unavailable", "error", err)
		return nil
	}

	// Hard-end path: EntStore.End invokes this processor with a bounded timeout.
	entStore.SetSessionEndProcessor(func(ctx context.Context, key string) error {
		return idx.IndexSession(ctx, key)
	})

	// Startup sweep to reprocess sessions left pending from a previous run
	// (crash, TUI hard-quit that timed out, soft-end from channel idle).
	app.registry.Register(lifecycle.NewFuncComponent("session-recall-sweep",
		func(ctx context.Context, wg *sync.WaitGroup) error {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := idx.ProcessPending(ctx); err != nil {
					logger().Warnw("session recall startup sweep failed", "error", err)
				}
			}()
			return nil
		},
		nil,
	), lifecycle.PriorityBuffer)

	return idx
}

// recallProviderAdapter adapts *session.RecallIndex to adk.RecallProvider
// using the configured topN and minRank.
type recallProviderAdapter struct {
	idx     *session.RecallIndex
	topN    int
	minRank float64
}

func newRecallProviderAdapter(idx *session.RecallIndex, cfg config.ContextRecallConfig) *recallProviderAdapter {
	resolved := cfg.ResolveRecall()
	return &recallProviderAdapter{
		idx:     idx,
		topN:    resolved.TopN,
		minRank: resolved.MinRank,
	}
}

func (a *recallProviderAdapter) RecallRecent(ctx context.Context, currentSessionKey, query string) ([]adk.RecallMatch, error) {
	if a.idx == nil {
		return nil, nil
	}
	results, err := a.idx.Search(ctx, query, a.topN*2) // over-fetch to survive self-filter.
	if err != nil {
		return nil, err
	}
	out := make([]adk.RecallMatch, 0, a.topN)
	for _, r := range results {
		if r.RowID == currentSessionKey {
			continue
		}
		// FTS5 rank is a negative BM25 score (lower = better match). Convert
		// to a positive comparability value: rankFloor is applied against
		// |rank| so users can reason about minRank as "a stronger signal
		// passes a higher floor."
		score := -r.Rank
		if score < a.minRank {
			continue
		}
		summary, _ := a.idx.GetSummary(ctx, r.RowID)
		out = append(out, adk.RecallMatch{
			SessionKey: r.RowID,
			Summary:    summary,
			Rank:       score,
		})
		if len(out) >= a.topN {
			break
		}
	}
	return out, nil
}
