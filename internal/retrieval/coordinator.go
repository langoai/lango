package retrieval

import (
	"context"
	"sort"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/langoai/lango/internal/knowledge"
	"github.com/langoai/lango/internal/types"
)

// RetrievalCoordinator runs multiple RetrievalAgents in parallel and merges results.
type RetrievalCoordinator struct {
	agents []RetrievalAgent
	logger *zap.SugaredLogger
	shadow bool // when true, results are for comparison only
}

// NewRetrievalCoordinator creates a coordinator with the given agents.
func NewRetrievalCoordinator(agents []RetrievalAgent, logger *zap.SugaredLogger) *RetrievalCoordinator {
	return &RetrievalCoordinator{
		agents: agents,
		logger: logger,
		shadow: true, // default: shadow mode
	}
}

// SetShadow toggles shadow mode where results are used for comparison logging only.
func (c *RetrievalCoordinator) SetShadow(shadow bool) {
	c.shadow = shadow
}

// Shadow reports whether the coordinator is in shadow mode.
func (c *RetrievalCoordinator) Shadow() bool {
	return c.shadow
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
	var mu sync.Mutex
	var allFindings []Finding

	g, gctx := errgroup.WithContext(ctx)
	for _, agent := range c.agents {
		g.Go(func() error {
			findings, err := agent.Search(gctx, query, defaultAgentLimit)
			if err != nil {
				c.logger.Warnw("retrieval agent error", "agent", agent.Name(), "error", err)
				return nil // non-fatal: continue with other agents
			}
			mu.Lock()
			allFindings = append(allFindings, findings...)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	deduped := dedupFindings(allFindings)
	sorted := sortFindingsByScore(deduped)

	if tokenBudget > 0 {
		sorted = TruncateFindings(sorted, tokenBudget)
	}

	return sorted, nil
}

// dedupFindings removes duplicate (Layer, Key) entries, keeping the highest Score.
func dedupFindings(findings []Finding) []Finding {
	best := make(map[dedupKey]Finding, len(findings))
	for _, f := range findings {
		dk := dedupKey{Layer: f.Layer, Key: f.Key}
		if existing, ok := best[dk]; !ok || f.Score > existing.Score {
			best[dk] = f
		}
	}

	result := make([]Finding, 0, len(best))
	for _, f := range best {
		result = append(result, f)
	}
	return result
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
