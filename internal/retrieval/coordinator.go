package retrieval

import (
	"context"
	"sort"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/types"
)

// RetrievalCoordinator runs multiple RetrievalAgents in parallel and merges results.
type RetrievalCoordinator struct {
	agents []RetrievalAgent
	logger *zap.SugaredLogger
}

// NewRetrievalCoordinator creates a coordinator with the given agents.
func NewRetrievalCoordinator(agents []RetrievalAgent, logger *zap.SugaredLogger) *RetrievalCoordinator {
	return &RetrievalCoordinator{
		agents: agents,
		logger: logger,
	}
}

// dedupKey is used to identify duplicate findings across agents.
type dedupKey struct {
	Layer knowledge.ContextLayer
	Key   string
}

// defaultAgentLimit is the per-agent item count limit passed to Search.
const defaultAgentLimit = 10

// Retrieve runs all agents in parallel, deduplicates, sorts, and optionally truncates findings.
func (c *RetrievalCoordinator) Retrieve(ctx context.Context, query string, tokenBudget int) ([]Finding, error) {
	results := make([][]Finding, len(c.agents))

	g, gctx := errgroup.WithContext(ctx)
	for i, agent := range c.agents {
		g.Go(func() error {
			findings, err := agent.Search(gctx, query, defaultAgentLimit)
			if err != nil {
				c.logger.Warnw("retrieval agent error", "agent", agent.Name(), "error", err)
				return nil // non-fatal: continue with other agents
			}
			results[i] = findings // each goroutine owns its index
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	var allFindings []Finding
	for _, r := range results {
		allFindings = append(allFindings, r...)
	}

	merged := mergeFindings(allFindings)
	sorted := sortFindingsByScore(merged)

	if tokenBudget > 0 {
		sorted = TruncateFindings(sorted, tokenBudget)
	}

	return sorted, nil
}

// sourceAuthority ranks knowledge authorship for evidence-based merge.
// Higher value = more authoritative. Unknown/empty source = 0.
var sourceAuthority = map[string]int{
	"knowledge":             4, // User-explicit via save_knowledge tool
	"session_learning":      3, // High-confidence end-of-session analysis
	"proactive_librarian":   2, // Auto-extracted / inquiry-confirmed
	"conversation_analysis": 1, // Real-time extraction (varied confidence)
	"memory":                1, // Observations/reflections
	"learning":              1, // Error pattern learning
}

// mergeFindings resolves duplicate (Layer, Key) entries using evidence-based
// priority: authority → version (supersedes) → recency → score.
// This replaces the previous score-only dedup. When all provenance fields are
// empty (e.g., ContextSearchAgent findings), the merge falls through to Score,
// preserving backward-compatible behavior.
func mergeFindings(findings []Finding) []Finding {
	best := make(map[dedupKey]Finding, len(findings))
	for _, f := range findings {
		dk := dedupKey{Layer: f.Layer, Key: f.Key}
		existing, ok := best[dk]
		if !ok || compareFindingPriority(f, existing) > 0 {
			best[dk] = f
		}
	}

	result := make([]Finding, 0, len(best))
	for _, f := range best {
		result = append(result, f)
	}
	return result
}

// compareFindingPriority returns >0 if a should be preferred over b,
// <0 if b is preferred, 0 if equal priority.
// Priority chain: authority → version (supersedes) → recency → score.
func compareFindingPriority(a, b Finding) int {
	// 1. Authority: higher source authority wins.
	authA := sourceAuthority[a.Source]
	authB := sourceAuthority[b.Source]
	if authA != authB {
		return authA - authB
	}

	// 2. Version: higher version supersedes lower.
	if a.Version != b.Version {
		return a.Version - b.Version
	}

	// 3. Recency: more recent UpdatedAt wins.
	if !a.UpdatedAt.IsZero() && !b.UpdatedAt.IsZero() {
		if a.UpdatedAt.After(b.UpdatedAt) {
			return 1
		}
		if b.UpdatedAt.After(a.UpdatedAt) {
			return -1
		}
	} else if !a.UpdatedAt.IsZero() {
		return 1
	} else if !b.UpdatedAt.IsZero() {
		return -1
	}

	// 4. Score: search relevance as final tiebreaker.
	if a.Score > b.Score {
		return 1
	}
	if b.Score > a.Score {
		return -1
	}

	return 0
}

// sortFindingsByScore sorts findings by Score descending.
func sortFindingsByScore(findings []Finding) []Finding {
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Score != findings[j].Score {
			return findings[i].Score > findings[j].Score
		}
		// Stable tie-breaking by key
		return findings[i].Key < findings[j].Key
	})
	return findings
}

// TruncateFindings drops lowest-score findings until total tokens fit within budget.
func TruncateFindings(findings []Finding, tokenBudget int) []Finding {
	if tokenBudget <= 0 {
		return findings
	}

	// findings are already sorted by score desc
	var total int
	for i, f := range findings {
		tokens := types.EstimateTokens(f.Content)
		if total+tokens > tokenBudget {
			return findings[:i]
		}
		total += tokens
	}
	return findings
}

// ToRetrievalResult converts findings into the knowledge.RetrievalResult format.
func ToRetrievalResult(findings []Finding) *knowledge.RetrievalResult {
	result := &knowledge.RetrievalResult{
		Items: make(map[knowledge.ContextLayer][]knowledge.ContextItem),
	}

	for _, f := range findings {
		item := knowledge.ContextItem{
			Layer:    f.Layer,
			Key:      f.Key,
			Content:  f.Content,
			Score:    f.Score,
			Source:   f.SearchSource,
			Category: f.Category,
		}
		result.Items[f.Layer] = append(result.Items[f.Layer], item)
		result.TotalItems++
	}

	return result
}
