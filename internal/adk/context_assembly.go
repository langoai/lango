package adk

import (
	"context"
	"strings"
	"sync"

	"github.com/langoai/lango/internal/memory"
)

// defaultMemoryTokenBudget is the default token budget for the memory section.
const defaultMemoryTokenBudget = 4000

// assembleMemorySection builds the "Conversation Memory" section from observations and reflections.
// It enforces a token budget: reflections are included first (higher information density),
// then observations fill the remaining budget.
func (m *ContextAwareModelAdapter) assembleMemorySection(ctx context.Context, sessionKey string) string {
	var reflections []memory.Reflection
	var observations []memory.Observation
	var err error

	if m.maxReflections > 0 {
		reflections, err = m.memoryProvider.ListRecentReflections(ctx, sessionKey, m.maxReflections)
	} else {
		reflections, err = m.memoryProvider.ListReflections(ctx, sessionKey)
	}
	if err != nil {
		m.logger.Warnw("memory reflection retrieval error", "error", err)
	}

	if m.maxObservations > 0 {
		observations, err = m.memoryProvider.ListRecentObservations(ctx, sessionKey, m.maxObservations)
	} else {
		observations, err = m.memoryProvider.ListObservations(ctx, sessionKey)
	}
	if err != nil {
		m.logger.Warnw("memory observation retrieval error", "error", err)
	}

	if len(reflections) == 0 && len(observations) == 0 {
		return ""
	}

	budget := m.memoryTokenBudget
	if budget <= 0 {
		budget = defaultMemoryTokenBudget
	}

	var b strings.Builder
	currentTokens := 0

	b.WriteString("## Conversation Memory\n")

	// Reflections first — higher information density from compressed summaries.
	if len(reflections) > 0 {
		b.WriteString("\n### Summary\n")
		for _, ref := range reflections {
			t := memory.EstimateTokens(ref.Content)
			if currentTokens+t > budget {
				break
			}
			b.WriteString(ref.Content)
			b.WriteString("\n")
			currentTokens += t
		}
	}

	// Observations fill remaining budget.
	if len(observations) > 0 && currentTokens < budget {
		b.WriteString("\n### Recent Observations\n")
		for _, obs := range observations {
			t := memory.EstimateTokens(obs.Content)
			if currentTokens+t > budget {
				break
			}
			b.WriteString("- ")
			b.WriteString(obs.Content)
			b.WriteString("\n")
			currentTokens += t
		}
	}

	return b.String()
}

func (m *ContextAwareModelAdapter) assembleRunSummarySection(ctx context.Context, sessionKey string) string {
	if m.runSummaryProvider == nil {
		return ""
	}

	maxSeq, err := m.runSummaryProvider.MaxJournalSeqForSession(ctx, sessionKey)
	if err != nil {
		m.logger.Warnw("run summary max seq retrieval error", "error", err)
		return ""
	}
	if cached, ok := m.getCachedRunSummary(sessionKey, maxSeq); ok {
		return cached
	}

	summaries, err := m.runSummaryProvider.ListRunSummaries(ctx, sessionKey, 3)
	if err != nil {
		m.logger.Warnw("run summary retrieval error", "error", err)
		return ""
	}
	if len(summaries) == 0 {
		m.storeCachedRunSummary(sessionKey, maxSeq, "")
		return ""
	}

	var b strings.Builder
	b.WriteString("## Active Runs\n")
	for _, summary := range summaries {
		b.WriteString("- ")
		b.WriteString(summary.RunID)
		b.WriteString(": ")
		b.WriteString(summary.Goal)
		b.WriteString(" [status=")
		b.WriteString(summary.Status)
		b.WriteString("]")
		if summary.CurrentStep != "" {
			b.WriteString(" current=")
			b.WriteString(summary.CurrentStep)
		}
		if summary.CurrentBlocker != "" {
			b.WriteString(" blocker=")
			b.WriteString(summary.CurrentBlocker)
		}
		b.WriteString("\n")
	}
	assembled := b.String()
	m.storeCachedRunSummary(sessionKey, maxSeq, assembled)
	return assembled
}

type runSummaryCache struct {
	mu      sync.RWMutex
	entries map[string]summaryCacheEntry
}

type summaryCacheEntry struct {
	summary string
	maxSeq  int64
}

func (m *ContextAwareModelAdapter) getCachedRunSummary(sessionKey string, maxSeq int64) (string, bool) {
	if m.runSummaryCache == nil {
		return "", false
	}
	m.runSummaryCache.mu.RLock()
	defer m.runSummaryCache.mu.RUnlock()

	entry, ok := m.runSummaryCache.entries[sessionKey]
	if !ok || entry.maxSeq != maxSeq {
		return "", false
	}
	return entry.summary, true
}

func (m *ContextAwareModelAdapter) storeCachedRunSummary(
	sessionKey string,
	maxSeq int64,
	summary string,
) {
	if m.runSummaryCache == nil {
		return
	}
	m.runSummaryCache.mu.Lock()
	defer m.runSummaryCache.mu.Unlock()
	m.runSummaryCache.entries[sessionKey] = summaryCacheEntry{
		summary: summary,
		maxSeq:  maxSeq,
	}
}
