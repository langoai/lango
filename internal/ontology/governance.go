package ontology

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// validTransitions defines the allowed FSM transitions.
// Key: from status, Value: set of allowed target statuses.
// immutable after init — do not modify at runtime.
var validTransitions = map[SchemaStatus]map[SchemaStatus]bool{
	SchemaProposed:    {SchemaShadow: true, SchemaQuarantined: true},
	SchemaQuarantined: {SchemaProposed: true},
	SchemaShadow:      {SchemaActive: true, SchemaQuarantined: true},
	SchemaActive:      {SchemaDeprecated: true},
	// SchemaDeprecated: no outgoing transitions (terminal)
}

// GovernanceEngine manages schema lifecycle FSM and rate limiting.
// Type/predicate agnostic — FSM rules and rate limits apply equally.
type GovernanceEngine struct {
	policy GovernancePolicy
	mu     sync.Mutex
	// In-memory daily proposal counter (resets on restart — acceptable for v1).
	dailyCounts map[string]int // date string "2006-01-02" → count
}

// NewGovernanceEngine creates a GovernanceEngine with the given policy.
func NewGovernanceEngine(policy GovernancePolicy) *GovernanceEngine {
	return &GovernanceEngine{
		policy:      policy,
		dailyCounts: make(map[string]int),
	}
}

// ValidateTransition checks whether a transition from → to is allowed by the FSM.
func (g *GovernanceEngine) ValidateTransition(from, to SchemaStatus) error {
	targets, ok := validTransitions[from]
	if !ok || !targets[to] {
		return fmt.Errorf("invalid schema transition: %s → %s", from, to)
	}
	return nil
}

// CheckRateLimit enforces the daily proposal limit (type + predicate combined).
func (g *GovernanceEngine) CheckRateLimit(_ context.Context) error {
	if g.policy.MaxNewPerDay <= 0 {
		return nil // no limit configured
	}
	g.mu.Lock()
	defer g.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	if g.dailyCounts[today] >= g.policy.MaxNewPerDay {
		return fmt.Errorf("daily schema proposal limit reached (%d/%d)", g.dailyCounts[today], g.policy.MaxNewPerDay)
	}
	g.dailyCounts[today]++
	return nil
}

// SchemaHealth returns status counts for types and predicates.
func (g *GovernanceEngine) SchemaHealth(ctx context.Context, registry Registry) (*SchemaHealthReport, error) {
	types, err := registry.ListTypes(ctx)
	if err != nil {
		return nil, fmt.Errorf("list types: %w", err)
	}
	preds, err := registry.ListPredicates(ctx)
	if err != nil {
		return nil, fmt.Errorf("list predicates: %w", err)
	}

	report := &SchemaHealthReport{
		Types:      make(map[SchemaStatus]int),
		Predicates: make(map[SchemaStatus]int),
	}
	for _, t := range types {
		report.Types[t.Status]++
	}
	for _, p := range preds {
		report.Predicates[p.Status]++
	}
	return report, nil
}

// TypeUsage returns basic info about a type's status and age.
// Full usage counting (triple/property counts) is deferred to a future observability change.
func (g *GovernanceEngine) TypeUsage(ctx context.Context, registry Registry, typeName string) (*TypeUsageInfo, error) {
	t, err := registry.GetType(ctx, typeName)
	if err != nil {
		return nil, fmt.Errorf("get type %q: %w", typeName, err)
	}
	return &TypeUsageInfo{
		TypeName:  t.Name,
		Status:    t.Status,
		Version:   t.Version,
		CreatedAt: t.CreatedAt,
	}, nil
}
