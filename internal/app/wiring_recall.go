package app

import (
	"context"
	"sync"
	"time"

	"github.com/langoai/lango/internal/adk"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/lifecycle"
	"github.com/langoai/lango/internal/search"
	"github.com/langoai/lango/internal/session"
)

type recallBackend interface {
	Search(ctx context.Context, query string, limit int) ([]search.SearchResult, error)
	GetSummary(ctx context.Context, key string) (string, error)
	ProcessPending(ctx context.Context) error
	IndexSession(ctx context.Context, key string) error
}

type recallSessionProcessorSetter interface {
	SetSessionEndProcessor(session.SessionEndProcessor)
	SetHardEndTimeout(time.Duration)
}

// wireSessionRecall constructs the recall backend, registers its startup sweep,
// and installs it as the session-end processor for hard-end flows.
func wireSessionRecall(app *App) recallBackend {
	if app == nil || app.Store == nil {
		return nil
	}
	cfg := app.Config.Context.Recall.ResolveRecall()
	if cfg.Enabled == nil || !*cfg.Enabled {
		return nil
	}

	var idx recallBackend
	if entStore, ok := app.Store.(*session.EntStore); ok {
		if entStore.DB() == nil {
			return nil
		}
		recallIdx := session.NewRecallIndex(entStore)
		if err := recallIdx.EnsureReady(); err != nil {
			logger().Warnw("session recall disabled — FTS5 unavailable", "error", err)
			return nil
		}
		idx = recallIdx
	} else if brokerStore, ok := app.Store.(interface {
		recallSessionProcessorSetter
		IndexSession(ctx context.Context, key string) error
		ProcessPending(ctx context.Context) error
		Search(ctx context.Context, query string, limit int) ([]search.SearchResult, error)
		GetSummary(ctx context.Context, key string) (string, error)
	}); ok {
		idx = brokerStore
	} else {
		return nil
	}

	if setter, ok := app.Store.(recallSessionProcessorSetter); ok {
		setter.SetSessionEndProcessor(func(ctx context.Context, key string) error {
			return idx.IndexSession(ctx, key)
		})
	}

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

// recallProviderAdapter adapts the recall backend to adk.RecallProvider using
// the configured topN and minRank.
type recallProviderAdapter struct {
	idx     recallBackend
	topN    int
	minRank float64
}

func newRecallProviderAdapter(idx recallBackend, cfg config.ContextRecallConfig) *recallProviderAdapter {
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
	results, err := a.idx.Search(ctx, query, a.topN*2)
	if err != nil {
		return nil, err
	}
	out := make([]adk.RecallMatch, 0, a.topN)
	for _, r := range results {
		if r.RowID == currentSessionKey {
			continue
		}
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
