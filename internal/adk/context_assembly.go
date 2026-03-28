package adk

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/langoai/lango/internal/memory"
	"github.com/langoai/lango/internal/types"
)

// defaultMemoryTokenBudget is the default token budget for the memory section.
const defaultMemoryTokenBudget = 4000

// retrieveMemoryData fetches reflections and observations for the session.
// Item count limits (maxReflections, maxObservations) are enforced here.
func (m *ContextAwareModelAdapter) retrieveMemoryData(ctx context.Context, sessionKey string) ([]memory.Reflection, []memory.Observation) {
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

	return reflections, observations
}

// formatMemorySection builds the "Conversation Memory" section from pre-retrieved data.
// Reflections are included first (higher information density), then observations fill
// the remaining budget. budget=0 means use default (4000).
func (m *ContextAwareModelAdapter) formatMemorySection(reflections []memory.Reflection, observations []memory.Observation, budget int) string {
	if len(reflections) == 0 && len(observations) == 0 {
		return ""
	}

	if budget <= 0 {
		budget = m.memoryTokenBudget
	}
	if budget <= 0 {
		budget = defaultMemoryTokenBudget
	}

	var b strings.Builder
	currentTokens := 0

	b.WriteString("## Conversation Memory\n")

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

// retrieveRunSummaryData fetches run summaries for the session.
// Returns nil if provider is absent, an error occurs, or no summaries exist.
// Uses a cache keyed by (sessionKey, maxJournalSeq) to avoid repeated queries.
func (m *ContextAwareModelAdapter) retrieveRunSummaryData(ctx context.Context, sessionKey string) []RunSummaryContext {
	if m.runSummaryProvider == nil {
		return nil
	}

	maxSeq, err := m.runSummaryProvider.MaxJournalSeqForSession(ctx, sessionKey)
	if err != nil {
		m.logger.Warnw("run summary max seq retrieval error", "error", err)
		return nil
	}
	if cached, ok := m.getCachedRunSummary(sessionKey, maxSeq); ok {
		return cached
	}

	summaries, err := m.runSummaryProvider.ListRunSummaries(ctx, sessionKey, 3)
	if err != nil {
		m.logger.Warnw("run summary retrieval error", "error", err)
		return nil
	}
	m.storeCachedRunSummary(sessionKey, maxSeq, summaries)
	return summaries
}

// formatRunSummarySection builds the "Active Runs" section from pre-retrieved summaries.
// budgetTokens controls item-level truncation: 0 = unlimited.
func formatRunSummarySection(summaries []RunSummaryContext, budgetTokens int) string {
	if len(summaries) == 0 {
		return ""
	}

	// Item-level truncation: drop older summaries until within budget.
	if budgetTokens > 0 {
		headerTokens := types.EstimateTokens("## Active Runs\n")
		remaining := budgetTokens - headerTokens
		for i, summary := range summaries {
			line := fmt.Sprintf("- %s: %s [status=%s]\n", summary.RunID, summary.Goal, summary.Status)
			itemTokens := types.EstimateTokens(line)
			if remaining-itemTokens < 0 {
				summaries = summaries[:i]
				break
			}
			remaining -= itemTokens
		}
	}

	if len(summaries) == 0 {
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
	return b.String()
}

type runSummaryCache struct {
	mu      sync.RWMutex
	entries map[string]summaryCacheEntry
}

type summaryCacheEntry struct {
	summaries []RunSummaryContext
	maxSeq    int64
}

func (m *ContextAwareModelAdapter) getCachedRunSummary(sessionKey string, maxSeq int64) ([]RunSummaryContext, bool) {
	if m.runSummaryCache == nil {
		return nil, false
	}
	m.runSummaryCache.mu.RLock()
	defer m.runSummaryCache.mu.RUnlock()

	entry, ok := m.runSummaryCache.entries[sessionKey]
	if !ok || entry.maxSeq != maxSeq {
		return nil, false
	}
	return entry.summaries, true
}

func (m *ContextAwareModelAdapter) storeCachedRunSummary(
	sessionKey string,
	maxSeq int64,
	summaries []RunSummaryContext,
) {
	if m.runSummaryCache == nil {
		return
	}
	m.runSummaryCache.mu.Lock()
	defer m.runSummaryCache.mu.Unlock()
	m.runSummaryCache.entries[sessionKey] = summaryCacheEntry{
		summaries: summaries,
		maxSeq:    maxSeq,
	}
}
