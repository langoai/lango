package adk

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/prompt"
	"github.com/langoai/lango/internal/provider"
)

func BenchmarkAssembleRunSummary_CacheHit(b *testing.B) {
	provider := &mockRunSummaryProvider{
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
	adapter.WithRunSummaryProvider(provider)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = adapter.assembleRunSummarySection(ctx, "sess-bench", 0)
	}
}

func BenchmarkAssembleRunSummary_CacheMiss(b *testing.B) {
	provider := &mockRunSummaryProvider{
		summaries: []RunSummaryContext{{
			RunID:          "run-1",
			Goal:           "Optimize cache",
			Status:         "running",
			CurrentStep:    "Summarize",
			CurrentBlocker: "none",
		}},
	}
	adapter := newBenchmarkContextAdapter()
	adapter.WithRunSummaryProvider(provider)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.maxSeq = int64(i + 1)
		_ = adapter.assembleRunSummarySection(ctx, "sess-bench", 0)
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
