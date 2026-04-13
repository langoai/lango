package streamx

import (
	"context"
	"fmt"
	"time"

	"github.com/langoai/lango/internal/agent"
	"golang.org/x/sync/errgroup"
)

// ToolInvocation represents a single tool call request.
type ToolInvocation struct {
	Tool   *agent.Tool
	Params map[string]any
}

// ToolResult represents the outcome of a tool invocation.
type ToolResult struct {
	ToolName string
	Result   any
	Error    error
	Duration time.Duration
}

// ParallelReadOnlyExecutor executes eligible read-only, concurrency-safe tools concurrently.
type ParallelReadOnlyExecutor struct {
	maxConcurrency int
}

// NewParallelReadOnlyExecutor creates a new executor with the given concurrency limit.
// If maxConcurrency is less than 1, it defaults to 1.
func NewParallelReadOnlyExecutor(maxConcurrency int) *ParallelReadOnlyExecutor {
	if maxConcurrency < 1 {
		maxConcurrency = 1
	}
	return &ParallelReadOnlyExecutor{maxConcurrency: maxConcurrency}
}

// IsEligible checks if a tool can be executed in parallel.
// A tool is eligible when both Capability.ReadOnly and Capability.ConcurrencySafe are true.
func IsEligible(t *agent.Tool) bool {
	if t == nil {
		return false
	}
	return t.Capability.ReadOnly && t.Capability.ConcurrencySafe
}

// ExecuteParallel runs eligible tools concurrently, respecting the configured
// concurrency limit. Non-eligible tools are rejected with an error in the result
// without being executed. Results are returned in the same order as invocations.
// Context cancellation stops pending invocations.
func (e *ParallelReadOnlyExecutor) ExecuteParallel(ctx context.Context, invocations []ToolInvocation) []ToolResult {
	results := make([]ToolResult, len(invocations))

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(e.maxConcurrency)

	for i, inv := range invocations {
		if inv.Tool == nil {
			results[i] = ToolResult{
				Error: fmt.Errorf("nil tool at index %d", i),
			}
			continue
		}

		if !IsEligible(inv.Tool) {
			results[i] = ToolResult{
				ToolName: inv.Tool.Name,
				Error: fmt.Errorf(
					"not eligible for parallel execution: ReadOnly=%v, ConcurrencySafe=%v",
					inv.Tool.Capability.ReadOnly,
					inv.Tool.Capability.ConcurrencySafe,
				),
			}
			continue
		}

		g.Go(func() error {
			// Check context before starting execution.
			if err := gctx.Err(); err != nil {
				results[i] = ToolResult{
					ToolName: inv.Tool.Name,
					Error:    err,
				}
				return nil // don't cancel the group; just record the error
			}

			start := time.Now()
			result, err := inv.Tool.Handler(gctx, inv.Params)
			duration := time.Since(start)

			results[i] = ToolResult{
				ToolName: inv.Tool.Name,
				Result:   result,
				Error:    err,
				Duration: duration,
			}
			return nil // errors are captured per-result, not propagated to the group
		})
	}

	// Wait never returns an error because goroutines always return nil.
	_ = g.Wait()

	return results
}
