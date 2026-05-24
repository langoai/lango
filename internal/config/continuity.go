package config

import (
	"log/slog"
	"time"
)

// Default values for Phase 3 continuity configs. Re-exported as constants so
// wiring callers can reference them without importing the loader.
const (
	DefaultCompactionThreshold   = 0.5
	DefaultCompactionSyncTimeout = 2 * time.Second
	DefaultCompactionWorkerCount = 1

	DefaultRecallTopN    = 3
	DefaultRecallMinRank = 0.2

	DefaultLearningSuggestionThreshold   = 0.5
	DefaultLearningSuggestionRateLimit   = 10
	DefaultLearningSuggestionDedupWindow = time.Hour
)

// ResolveCompaction clamps the compaction config to valid ranges, logging a
// warning for each clamped field. The returned struct always has a non-nil
// Enabled pointer.
func (c ContextCompactionConfig) ResolveCompaction() ContextCompactionConfig {
	out := c
	if out.Enabled == nil {
		t := true
		out.Enabled = &t
	}
	if out.Threshold == 0 {
		out.Threshold = DefaultCompactionThreshold
	} else if out.Threshold < 0.1 {
		slog.Warn("context.compaction.threshold below min; clamping", "value", out.Threshold, "min", 0.1)
		out.Threshold = 0.1
	} else if out.Threshold > 0.95 {
		slog.Warn("context.compaction.threshold above max; clamping", "value", out.Threshold, "max", 0.95)
		out.Threshold = 0.95
	}
	if out.SyncTimeout == 0 {
		out.SyncTimeout = DefaultCompactionSyncTimeout
	} else if out.SyncTimeout < 100*time.Millisecond {
		slog.Warn("context.compaction.syncTimeout below min; clamping", "value", out.SyncTimeout, "min", 100*time.Millisecond)
		out.SyncTimeout = 100 * time.Millisecond
	} else if out.SyncTimeout > 10*time.Second {
		slog.Warn("context.compaction.syncTimeout above max; clamping", "value", out.SyncTimeout, "max", 10*time.Second)
		out.SyncTimeout = 10 * time.Second
	}
	if out.WorkerCount <= 0 {
		out.WorkerCount = DefaultCompactionWorkerCount
	}
	return out
}

// ResolveRecall clamps the recall config to valid ranges, logging a warning
// for each clamped field. The returned struct always has a non-nil Enabled
// pointer.
func (c ContextRecallConfig) ResolveRecall() ContextRecallConfig {
	out := c
	if out.Enabled == nil {
		t := true
		out.Enabled = &t
	}
	if out.TopN == 0 {
		out.TopN = DefaultRecallTopN
	} else if out.TopN < 1 {
		slog.Warn("context.recall.topN below min; clamping", "value", out.TopN, "min", 1)
		out.TopN = 1
	} else if out.TopN > 10 {
		slog.Warn("context.recall.topN above max; clamping", "value", out.TopN, "max", 10)
		out.TopN = 10
	}
	if out.MinRank == 0 {
		out.MinRank = DefaultRecallMinRank
	} else if out.MinRank < 0.0 {
		slog.Warn("context.recall.minRank below min; clamping", "value", out.MinRank, "min", 0.0)
		out.MinRank = 0.0
	} else if out.MinRank > 1.0 {
		slog.Warn("context.recall.minRank above max; clamping", "value", out.MinRank, "max", 1.0)
		out.MinRank = 1.0
	}
	return out
}

// ResolveSuggestions clamps the learning-suggestion config to valid ranges,
// logging a warning for each clamped field. The returned struct always has a
// non-nil Enabled pointer.
func (c LearningSuggestionsConfig) ResolveSuggestions() LearningSuggestionsConfig {
	out := c
	if out.Enabled == nil {
		t := true
		out.Enabled = &t
	}
	if out.Threshold == 0 {
		out.Threshold = DefaultLearningSuggestionThreshold
	} else if out.Threshold < 0.1 {
		slog.Warn("learning.suggestions.threshold below min; clamping", "value", out.Threshold, "min", 0.1)
		out.Threshold = 0.1
	} else if out.Threshold > 0.9 {
		slog.Warn("learning.suggestions.threshold above max; clamping", "value", out.Threshold, "max", 0.9)
		out.Threshold = 0.9
	}
	if out.RateLimit == 0 {
		out.RateLimit = DefaultLearningSuggestionRateLimit
	} else if out.RateLimit < 1 {
		slog.Warn("learning.suggestions.rateLimit below min; clamping", "value", out.RateLimit, "min", 1)
		out.RateLimit = 1
	} else if out.RateLimit > 100 {
		slog.Warn("learning.suggestions.rateLimit above max; clamping", "value", out.RateLimit, "max", 100)
		out.RateLimit = 100
	}
	if out.DedupWindow == 0 {
		out.DedupWindow = DefaultLearningSuggestionDedupWindow
	} else if out.DedupWindow < time.Minute {
		slog.Warn("learning.suggestions.dedupWindow below min; clamping", "value", out.DedupWindow, "min", time.Minute)
		out.DedupWindow = time.Minute
	} else if out.DedupWindow > 24*time.Hour {
		slog.Warn("learning.suggestions.dedupWindow above max; clamping", "value", out.DedupWindow, "max", 24*time.Hour)
		out.DedupWindow = 24 * time.Hour
	}
	return out
}
