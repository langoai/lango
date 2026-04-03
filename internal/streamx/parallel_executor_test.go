package streamx

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/langoai/lango/internal/agent"
)

// newEligibleTool creates a tool that is eligible for parallel execution.
func newEligibleTool(name string, handler agent.ToolHandler) *agent.Tool {
	return agent.NewTool(name, "test tool").
		ReadOnly().
		ConcurrencySafe().
		Handler(handler).
		Build()
}

// newIneligibleTool creates a tool that is NOT eligible for parallel execution.
func newIneligibleTool(name string, handler agent.ToolHandler) *agent.Tool {
	return agent.NewTool(name, "test tool").
		Handler(handler).
		Build()
}

func TestIsEligible(t *testing.T) {
	t.Parallel()

	t.Run("nil tool", func(t *testing.T) {
		t.Parallel()
		if IsEligible(nil) {
			t.Error("nil tool should not be eligible")
		}
	})

	t.Run("read-only and concurrency-safe", func(t *testing.T) {
		t.Parallel()
		tool := newEligibleTool("test", nil)
		if !IsEligible(tool) {
			t.Error("tool with ReadOnly=true and ConcurrencySafe=true should be eligible")
		}
	})

	t.Run("read-only only", func(t *testing.T) {
		t.Parallel()
		tool := agent.NewTool("test", "desc").ReadOnly().Build()
		if IsEligible(tool) {
			t.Error("tool with only ReadOnly=true should not be eligible")
		}
	})

	t.Run("concurrency-safe only", func(t *testing.T) {
		t.Parallel()
		tool := agent.NewTool("test", "desc").ConcurrencySafe().Build()
		if IsEligible(tool) {
			t.Error("tool with only ConcurrencySafe=true should not be eligible")
		}
	})

	t.Run("neither", func(t *testing.T) {
		t.Parallel()
		tool := agent.NewTool("test", "desc").Build()
		if IsEligible(tool) {
			t.Error("tool with neither flag should not be eligible")
		}
	})
}

func TestParallelReadOnlyExecutor_AllEligibleSucceed(t *testing.T) {
	t.Parallel()

	var running atomic.Int32
	var maxRunning atomic.Int32

	handler := func(_ context.Context, _ map[string]any) (any, error) {
		cur := running.Add(1)
		// Track peak concurrency.
		for {
			prev := maxRunning.Load()
			if cur <= prev || maxRunning.CompareAndSwap(prev, cur) {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
		running.Add(-1)
		return "ok", nil
	}

	invocations := []ToolInvocation{
		{Tool: newEligibleTool("tool-a", handler), Params: nil},
		{Tool: newEligibleTool("tool-b", handler), Params: nil},
		{Tool: newEligibleTool("tool-c", handler), Params: nil},
	}

	exec := NewParallelReadOnlyExecutor(10)
	results := exec.ExecuteParallel(context.Background(), invocations)

	if len(results) != 3 {
		t.Fatalf("want 3 results, got %d", len(results))
	}

	for i, r := range results {
		if r.Error != nil {
			t.Errorf("results[%d]: unexpected error: %v", i, r.Error)
		}
		if r.Result != "ok" {
			t.Errorf("results[%d]: want result 'ok', got %v", i, r.Result)
		}
		if r.ToolName != invocations[i].Tool.Name {
			t.Errorf("results[%d]: want tool name %q, got %q", i, invocations[i].Tool.Name, r.ToolName)
		}
		if r.Duration == 0 {
			t.Errorf("results[%d]: duration should be non-zero", i)
		}
	}

	// Verify tools actually ran in parallel.
	if peak := maxRunning.Load(); peak < 2 {
		t.Errorf("expected parallel execution (peak concurrency >= 2), got %d", peak)
	}
}

func TestParallelReadOnlyExecutor_MixedEligibility(t *testing.T) {
	t.Parallel()

	handler := func(_ context.Context, _ map[string]any) (any, error) {
		return "executed", nil
	}

	invocations := []ToolInvocation{
		{Tool: newEligibleTool("eligible-1", handler), Params: nil},
		{Tool: newIneligibleTool("ineligible-1", handler), Params: nil},
		{Tool: newEligibleTool("eligible-2", handler), Params: nil},
	}

	exec := NewParallelReadOnlyExecutor(4)
	results := exec.ExecuteParallel(context.Background(), invocations)

	if len(results) != 3 {
		t.Fatalf("want 3 results, got %d", len(results))
	}

	// Eligible tools succeed.
	if results[0].Error != nil {
		t.Errorf("results[0] (eligible): unexpected error: %v", results[0].Error)
	}
	if results[0].Result != "executed" {
		t.Errorf("results[0]: want 'executed', got %v", results[0].Result)
	}

	// Non-eligible tool gets error.
	if results[1].Error == nil {
		t.Fatal("results[1] (ineligible): expected error, got nil")
	}
	if results[1].ToolName != "ineligible-1" {
		t.Errorf("results[1]: want tool name 'ineligible-1', got %q", results[1].ToolName)
	}
	// Result should not be set for ineligible tools.
	if results[1].Result != nil {
		t.Errorf("results[1] (ineligible): result should be nil, got %v", results[1].Result)
	}

	// Second eligible tool succeeds.
	if results[2].Error != nil {
		t.Errorf("results[2] (eligible): unexpected error: %v", results[2].Error)
	}
	if results[2].Result != "executed" {
		t.Errorf("results[2]: want 'executed', got %v", results[2].Result)
	}
}

func TestParallelReadOnlyExecutor_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	started := make(chan struct{})
	handler := func(ctx context.Context, _ map[string]any) (any, error) {
		started <- struct{}{}
		<-ctx.Done()
		return nil, ctx.Err()
	}

	invocations := []ToolInvocation{
		{Tool: newEligibleTool("blocking", handler), Params: nil},
		{Tool: newEligibleTool("blocking-2", handler), Params: nil},
	}

	exec := NewParallelReadOnlyExecutor(2)

	done := make(chan []ToolResult, 1)
	go func() {
		done <- exec.ExecuteParallel(ctx, invocations)
	}()

	// Wait for at least one goroutine to start, then cancel.
	<-started
	cancel()

	results := <-done

	if len(results) != 2 {
		t.Fatalf("want 2 results, got %d", len(results))
	}

	// At least one result should have a context error.
	var gotCtxErr bool
	for _, r := range results {
		if r.Error != nil && errors.Is(r.Error, context.Canceled) {
			gotCtxErr = true
		}
	}
	if !gotCtxErr {
		t.Error("expected at least one context.Canceled error after cancellation")
	}
}

func TestParallelReadOnlyExecutor_MaxConcurrencyLimiting(t *testing.T) {
	t.Parallel()

	const maxConc = 2
	var running atomic.Int32
	var maxRunning atomic.Int32

	handler := func(_ context.Context, _ map[string]any) (any, error) {
		cur := running.Add(1)
		for {
			prev := maxRunning.Load()
			if cur <= prev || maxRunning.CompareAndSwap(prev, cur) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
		running.Add(-1)
		return "ok", nil
	}

	invocations := make([]ToolInvocation, 6)
	for i := range invocations {
		invocations[i] = ToolInvocation{
			Tool:   newEligibleTool("tool", handler),
			Params: nil,
		}
	}

	exec := NewParallelReadOnlyExecutor(maxConc)
	results := exec.ExecuteParallel(context.Background(), invocations)

	if len(results) != 6 {
		t.Fatalf("want 6 results, got %d", len(results))
	}

	for i, r := range results {
		if r.Error != nil {
			t.Errorf("results[%d]: unexpected error: %v", i, r.Error)
		}
	}

	if peak := maxRunning.Load(); peak > int32(maxConc) {
		t.Errorf("peak concurrency %d exceeded limit %d", peak, maxConc)
	}
}

func TestParallelReadOnlyExecutor_EmptyInvocations(t *testing.T) {
	t.Parallel()

	exec := NewParallelReadOnlyExecutor(4)
	results := exec.ExecuteParallel(context.Background(), nil)

	if len(results) != 0 {
		t.Fatalf("want 0 results, got %d", len(results))
	}

	results = exec.ExecuteParallel(context.Background(), []ToolInvocation{})
	if len(results) != 0 {
		t.Fatalf("want 0 results for empty slice, got %d", len(results))
	}
}

func TestParallelReadOnlyExecutor_SingleInvocation(t *testing.T) {
	t.Parallel()

	handler := func(_ context.Context, params map[string]any) (any, error) {
		return params["key"], nil
	}

	invocations := []ToolInvocation{
		{
			Tool:   newEligibleTool("single", handler),
			Params: map[string]any{"key": "value"},
		},
	}

	exec := NewParallelReadOnlyExecutor(4)
	results := exec.ExecuteParallel(context.Background(), invocations)

	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Error != nil {
		t.Fatalf("unexpected error: %v", r.Error)
	}
	if r.Result != "value" {
		t.Errorf("want result 'value', got %v", r.Result)
	}
	if r.ToolName != "single" {
		t.Errorf("want tool name 'single', got %q", r.ToolName)
	}
}

func TestParallelReadOnlyExecutor_HandlerError(t *testing.T) {
	t.Parallel()

	errBoom := errors.New("boom")
	handler := func(_ context.Context, _ map[string]any) (any, error) {
		return nil, errBoom
	}

	invocations := []ToolInvocation{
		{Tool: newEligibleTool("failing", handler), Params: nil},
	}

	exec := NewParallelReadOnlyExecutor(4)
	results := exec.ExecuteParallel(context.Background(), invocations)

	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}

	r := results[0]
	if !errors.Is(r.Error, errBoom) {
		t.Errorf("want error 'boom', got %v", r.Error)
	}
	if r.ToolName != "failing" {
		t.Errorf("want tool name 'failing', got %q", r.ToolName)
	}
	if r.Duration == 0 {
		t.Error("duration should be non-zero even for failing handler")
	}
}

func TestNewParallelReadOnlyExecutor_InvalidConcurrency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give int
		want int
	}{
		{give: 0, want: 1},
		{give: -1, want: 1},
		{give: 1, want: 1},
		{give: 10, want: 10},
	}

	for _, tt := range tests {
		exec := NewParallelReadOnlyExecutor(tt.give)
		if exec.maxConcurrency != tt.want {
			t.Errorf("NewParallelReadOnlyExecutor(%d): want maxConcurrency=%d, got %d", tt.give, tt.want, exec.maxConcurrency)
		}
	}
}

func TestParallelReadOnlyExecutor_ResultOrder(t *testing.T) {
	t.Parallel()

	// Each tool returns its own name to verify ordering.
	makeHandler := func(name string) agent.ToolHandler {
		return func(_ context.Context, _ map[string]any) (any, error) {
			// Add small jitter so execution order might differ from invocation order.
			time.Sleep(time.Duration(len(name)) * time.Millisecond)
			return name, nil
		}
	}

	invocations := []ToolInvocation{
		{Tool: newEligibleTool("alpha", makeHandler("alpha")), Params: nil},
		{Tool: newEligibleTool("beta", makeHandler("beta")), Params: nil},
		{Tool: newEligibleTool("gamma", makeHandler("gamma")), Params: nil},
		{Tool: newEligibleTool("delta", makeHandler("delta")), Params: nil},
	}

	exec := NewParallelReadOnlyExecutor(4)
	results := exec.ExecuteParallel(context.Background(), invocations)

	if len(results) != 4 {
		t.Fatalf("want 4 results, got %d", len(results))
	}

	// Results must be in invocation order regardless of completion order.
	wantNames := []string{"alpha", "beta", "gamma", "delta"}
	for i, want := range wantNames {
		if results[i].ToolName != want {
			t.Errorf("results[%d]: want tool name %q, got %q", i, want, results[i].ToolName)
		}
		if results[i].Result != want {
			t.Errorf("results[%d]: want result %q, got %v", i, want, results[i].Result)
		}
	}
}

func TestParallelReadOnlyExecutor_NilTool(t *testing.T) {
	t.Parallel()

	invocations := []ToolInvocation{
		{Tool: nil, Params: nil},
	}

	exec := NewParallelReadOnlyExecutor(4)
	results := exec.ExecuteParallel(context.Background(), invocations)

	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}

	if results[0].Error == nil {
		t.Fatal("expected error for nil tool, got nil")
	}
}
