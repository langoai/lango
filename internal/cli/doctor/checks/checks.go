// Package checks provides diagnostic check implementations for the doctor command.
package checks

import (
	"context"

	"github.com/langoai/lango/internal/config"
)

// Status represents the result status of a check.
type Status int

const (
	// StatusPass indicates the check passed successfully.
	StatusPass Status = iota
	// StatusWarn indicates the check passed with warnings.
	StatusWarn
	// StatusFail indicates the check failed.
	StatusFail
	// StatusSkip indicates the check was skipped.
	StatusSkip
)

// String returns a string representation of the status.
func (s Status) String() string {
	switch s {
	case StatusPass:
		return "pass"
	case StatusWarn:
		return "warn"
	case StatusFail:
		return "fail"
	case StatusSkip:
		return "skip"
	default:
		return "unknown"
	}
}

// Result represents the result of a single check.
type Result struct {
	// Name is the human-readable name of the check.
	Name string `json:"name"`
	// Status is the result status.
	Status Status `json:"status"`
	// Message is the main result message.
	Message string `json:"message"`
	// Details provides additional information or hints.
	Details string `json:"details,omitempty"`
	// Fixable indicates if this issue can be auto-repaired.
	Fixable bool `json:"fixable,omitempty"`
	// FixAction is a description of the fix action.
	FixAction string `json:"fixAction,omitempty"`
}

// Check is the interface that all diagnostic checks must implement.
type Check interface {
	// Name returns the human-readable name of the check.
	Name() string
	// Run executes the check and returns the result.
	Run(ctx context.Context, cfg *config.Config) Result
	// Fix attempts to repair the issue if possible.
	// Returns an updated result after the fix attempt.
	Fix(ctx context.Context, cfg *config.Config) Result
}

// Summary aggregates multiple check results.
type Summary struct {
	Results  []Result `json:"results"`
	Passed   int      `json:"passed"`
	Warnings int      `json:"warnings"`
	Failed   int      `json:"failed"`
	Skipped  int      `json:"skipped"`
}

// NewSummary creates a Summary from a slice of Results.
func NewSummary(results []Result) Summary {
	s := Summary{Results: results}
	for _, r := range results {
		switch r.Status {
		case StatusPass:
			s.Passed++
		case StatusWarn:
			s.Warnings++
		case StatusFail:
			s.Failed++
		case StatusSkip:
			s.Skipped++
		}
	}
	return s
}

// HasErrors returns true if any check failed.
func (s Summary) HasErrors() bool {
	return s.Failed > 0
}

// HasWarnings returns true if any check has warnings.
func (s Summary) HasWarnings() bool {
	return s.Warnings > 0
}

// AllChecks returns all available diagnostic checks.
func AllChecks() []Check {
	return []Check{
		&ConfigCheck{},
		&ProvidersCheck{},
		&APIKeySecurityCheck{},
		&ChannelCheck{},
		&DatabaseCheck{},
		&NetworkCheck{},
		// Security Checks
		&SecurityCheck{},
		&CompanionConnectionCheck{},
		// Memory & Scanning Checks
		&ObservationalMemoryCheck{},
		&OutputScanningCheck{},
		// Embedding / RAG
		&EmbeddingCheck{},
		// Graph / Multi-Agent / A2A
		&GraphStoreCheck{},
		&MultiAgentCheck{},
		&A2ACheck{},
	}
}
