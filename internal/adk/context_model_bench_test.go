package adk

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/prompt"
	"github.com/langoai/lango/internal/provider"
)

func BenchmarkRetrieveRunSummaryData_CacheHit(b *testing.B) {
	prov := &mockRunSummaryProvider{
		maxSeq: 1,
		summaries: []RunSummaryContext{{
			RunID:          "run-1",
			Goal:           "Optimize cache",
			Status:         "running",
			CurrentStep:    "Summarize",
			CurrentBlocker: "none",
		}},
	}
	adapter := newBenchmarkContextAdapter()
	adapter.WithRunSummaryProvider(prov)
	ctx := context.Background()

	// Prime the cache.
	_ = adapter.retrieveRunSummaryData(ctx, "sess-bench")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = adapter.retrieveRunSummaryData(ctx, "sess-bench")
	}
}

func BenchmarkRetrieveRunSummaryData_CacheMiss(b *testing.B) {
	prov := &mockRunSummaryProvider{
		summaries: []RunSummaryContext{{
			RunID:          "run-1",
			Goal:           "Optimize cache",
			Status:         "running",
			CurrentStep:    "Summarize",
			CurrentBlocker: "none",
		}},
	}
	adapter := newBenchmarkContextAdapter()
	adapter.WithRunSummaryProvider(prov)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prov.maxSeq = int64(i + 1)
		_ = adapter.retrieveRunSummaryData(ctx, "sess-bench")
	}
}

func BenchmarkFormatRunSummarySection(b *testing.B) {
	summaries := []RunSummaryContext{{
		RunID:          "run-1",
		Goal:           "Optimize cache",
		Status:         "running",
		CurrentStep:    "Summarize",
		CurrentBlocker: "none",
	}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = formatRunSummarySection(summaries, 0)
	}
}

func newBenchmarkContextAdapter() *ContextAwareModelAdapter {
	p := &mockProvider{
		id: "bench",
		events: []provider.StreamEvent{
			{Type: provider.StreamEventPlainText, Text: "ok"},
			{Type: provider.StreamEventDone},
		},
	}
	inner := NewModelAdapter(p, "bench-model")
	builder := prompt.DefaultBuilder()
	logger := zap.NewNop().Sugar()
	return NewContextAwareModelAdapter(inner, nil, builder, logger)
}
