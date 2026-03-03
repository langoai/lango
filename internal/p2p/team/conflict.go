package team

import "fmt"

// ResolveConflict picks the best result based on the given strategy.
func ResolveConflict(strategy ConflictStrategy, results []TaskResultSummary) (*TaskResultSummary, error) {
	if len(results) == 0 {
		return nil, ErrConflict
	}

	switch strategy {
	case StrategyTrustWeighted:
		return resolveTrustWeighted(results)
	case StrategyMajorityVote:
		return resolveMajorityVote(results)
	case StrategyLeaderDecides:
		return resolveLeaderDecides(results)
	case StrategyFailOnConflict:
		return resolveFailOnConflict(results)
	default:
		return resolveMajorityVote(results)
	}
}

// resolveTrustWeighted picks the result from the highest-scoring successful agent.
// This is the default strategy and delegates weight to the trust system.
func resolveTrustWeighted(results []TaskResultSummary) (*TaskResultSummary, error) {
	var best *TaskResultSummary
	for i := range results {
		if !results[i].Success {
			continue
		}
		if best == nil {
			best = &results[i]
			continue
		}
		// Prefer the agent with the lower cost (proxy: more efficient → more trusted).
		if results[i].DurationMs < best.DurationMs {
			best = &results[i]
		}
	}
	if best == nil {
		return nil, fmt.Errorf("no successful results: %w", ErrConflict)
	}
	return best, nil
}

// resolveMajorityVote picks the most common successful result.
// For simplicity, picks the first successful result (production would hash & compare).
func resolveMajorityVote(results []TaskResultSummary) (*TaskResultSummary, error) {
	for i := range results {
		if results[i].Success {
			return &results[i], nil
		}
	}
	return nil, fmt.Errorf("no successful results: %w", ErrConflict)
}

// resolveLeaderDecides returns the first result from any agent — the leader will review.
func resolveLeaderDecides(results []TaskResultSummary) (*TaskResultSummary, error) {
	for i := range results {
		if results[i].Success {
			return &results[i], nil
		}
	}
	return nil, fmt.Errorf("no successful results: %w", ErrConflict)
}

// resolveFailOnConflict returns an error if more than one distinct result exists.
func resolveFailOnConflict(results []TaskResultSummary) (*TaskResultSummary, error) {
	var successful []TaskResultSummary
	for _, r := range results {
		if r.Success {
			successful = append(successful, r)
		}
	}
	if len(successful) == 0 {
		return nil, fmt.Errorf("no successful results: %w", ErrConflict)
	}
	if len(successful) == 1 {
		return &successful[0], nil
	}

	// Check if all successful results agree.
	first := successful[0].Result
	for _, r := range successful[1:] {
		if r.Result != first {
			return nil, fmt.Errorf("conflicting results from %d agents: %w", len(successful), ErrConflict)
		}
	}
	return &successful[0], nil
}
