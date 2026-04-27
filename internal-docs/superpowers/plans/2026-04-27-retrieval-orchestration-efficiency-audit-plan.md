# Retrieval Orchestration Efficiency Audit Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Map retrieval/orchestration data flow and land behavior-preserving efficiency quick wins in the most obvious local hot paths.

**Architecture:** Use OpenSpec to document the work first, then make two narrow Go optimizations: pre-size retrieval aggregation containers and remove duplicate token estimation in context truncation. Keep retrieval ranking, evidence priority, Graph RAG semantics, and delegation behavior unchanged.

**Tech Stack:** Go 1.25, OpenSpec, `internal/knowledge`, `internal/retrieval`, `internal-docs/superpowers`, standard Go tests and benchmarks

---

## File Map

- Create: `openspec/changes/retrieval-orchestration-efficiency-audit/proposal.md`
  - Records why this efficiency audit exists and which surfaces are intentionally excluded.
- Create: `openspec/changes/retrieval-orchestration-efficiency-audit/design.md`
  - Contains the code-grounded flow map from request entry to retrieval, orchestration, and fan-in.
- Create: `openspec/changes/retrieval-orchestration-efficiency-audit/tasks.md`
  - Tracks audit, implementation, tests, verification, sync, and archive steps.
- Create: `openspec/changes/retrieval-orchestration-efficiency-audit/specs/retrieval-orchestration-efficiency/spec.md`
  - Defines behavior-preserving requirements for flow mapping and quick wins.
- Modify: `internal/retrieval/coordinator.go`
  - Pre-size aggregation slices and maps while preserving merge/sort/truncation behavior.
- Modify: `internal/retrieval/coordinator_test.go`
  - Add behavior-preserving tests and optional benchmark coverage for aggregation/truncation.
- Modify: `internal/knowledge/retriever.go`
  - Avoid duplicate token estimation in `TruncateResult` by computing per-item token cost once.
- Modify: `internal/knowledge/retriever_test.go`
  - Add tests proving priority order and truncation results remain unchanged.
- Modify only if implementation changes user-facing behavior: public docs under `docs/`
  - Expected result for this plan: no public docs changes.

## Task 1: Create OpenSpec Change Artifacts

**Files:**
- Create: `openspec/changes/retrieval-orchestration-efficiency-audit/proposal.md`
- Create: `openspec/changes/retrieval-orchestration-efficiency-audit/design.md`
- Create: `openspec/changes/retrieval-orchestration-efficiency-audit/tasks.md`
- Create: `openspec/changes/retrieval-orchestration-efficiency-audit/specs/retrieval-orchestration-efficiency/spec.md`

- [ ] **Step 1: Start the OpenSpec change**

Run through the repository OpenSpec workflow:

```text
/opsx:new retrieval-orchestration-efficiency-audit
/opsx:ff retrieval-orchestration-efficiency-audit
```

Expected:

```text
openspec/changes/retrieval-orchestration-efficiency-audit/ exists
proposal.md, design.md, tasks.md, and specs are ready to edit
```

- [ ] **Step 2: Fill the proposal**

Write `openspec/changes/retrieval-orchestration-efficiency-audit/proposal.md` with this content:

```markdown
# Proposal: Retrieval Orchestration Efficiency Audit

## Summary

Audit the boundary where context retrieval, agentic retrieval, Graph RAG, and multi-agent orchestration meet. Land only behavior-preserving quick wins that reduce avoidable repeated work.

## Motivation

Recent feature work expanded retrieval, memory, Graph RAG, structured agent runtime, and multi-agent orchestration. The code now has several paths where a user turn can trigger retrieval, merge, sort, token estimation, prompt assembly, and delegation bookkeeping. Before larger performance work, the system needs a code-grounded flow map and a small set of local optimizations.

## Non-Goals

- Change retrieval ranking or evidence priority.
- Replace Graph RAG traversal algorithms.
- Redesign agent runtime or P2P semantics.
- Redesign configuration.
- Expand public docs unless behavior becomes user-facing.

## Success Criteria

- A code-grounded retrieval/orchestration flow map exists in `design.md`.
- One or two local quick wins are implemented or explicitly rejected with evidence.
- Retrieval and context truncation behavior remains unchanged.
- `go build ./...` and `go test ./...` pass.
```

- [ ] **Step 3: Fill the delta spec**

Write `openspec/changes/retrieval-orchestration-efficiency-audit/specs/retrieval-orchestration-efficiency/spec.md` with this content:

```markdown
## Purpose

Behavior-preserving efficiency requirements for the retrieval and orchestration boundary.

## Requirements

### Requirement: Retrieval orchestration flow map
The change SHALL document the real code path from a user turn through context retrieval, agentic retrieval, Graph RAG, memory retrieval, orchestration delegation, and stream fan-in.

#### Scenario: Flow map identifies repeated work candidates
- **WHEN** the design artifact is reviewed
- **THEN** it lists where same-turn repeated query/search, merge, sort, token estimation, allocation, or fan-in bookkeeping can occur

### Requirement: Retrieval aggregation preallocation preserves behavior
`RetrievalCoordinator.Retrieve` and `mergeFindings` SHALL pre-size aggregation containers where cardinality is known without changing deduplication, score ordering, authority priority, or token truncation behavior.

#### Scenario: Authority merge still wins over score
- **WHEN** same-layer findings share a key but have different source authority
- **THEN** the higher-authority finding is retained even if its score is lower

#### Scenario: Different layers still preserve same key
- **WHEN** findings share a key across different context layers
- **THEN** both findings remain present after merge

### Requirement: Context truncation estimates each item once
`knowledge.TruncateResult` SHALL avoid estimating token count for the same item more than once in a single truncation call.

#### Scenario: Existing truncation behavior is preserved
- **WHEN** a result exceeds the budget
- **THEN** items are retained in the existing layer priority order until the budget is exhausted

#### Scenario: Too-small budget still returns zero items
- **WHEN** the first item exceeds the budget
- **THEN** the truncated result contains zero items
```

- [ ] **Step 4: Fill the OpenSpec design**

Write `openspec/changes/retrieval-orchestration-efficiency-audit/design.md` with this content:

```markdown
# Design: Retrieval Orchestration Efficiency Audit

## Code-Grounded Flow Map

1. User turn enters TurnRunner / ADK model adapter.
2. `ContextAwareModelAdapter` injects retrieved context before model execution.
3. `ContextRetriever.Retrieve` extracts keywords, walks requested layers, and calls knowledge, skills, external refs, learnings, tool registry, runtime context, and pending inquiries as configured.
4. Agentic retrieval uses `RetrievalCoordinator.Retrieve`, runs retrieval agents in parallel, appends all findings, merges by `(Layer, Key)`, sorts by score, and truncates by token budget.
5. Graph RAG runs a content retrieval phase and expands graph nodes when configured.
6. Observational memory can add observations and reflections to prompt context.
7. Structured orchestration routes tool-requiring work through `BuildAgentTree`, sub-agent routing entries, and `CoordinatingExecutor`.
8. Child agent output can be merged through `streamx.AgentStreamFanIn`.

## Quick Wins Selected

### Retrieval aggregation preallocation

`RetrievalCoordinator.Retrieve` already knows the number of retrieval-agent result buckets. After agent completion it can count total findings once and allocate `allFindings` with exact capacity. `mergeFindings` can keep its existing map capacity and return slice allocation. This preserves ordering because final ordering still comes from `sortFindingsByScore`.

### Context truncation token-cost cache

`knowledge.TruncateResult` currently estimates all item token costs to check whether truncation is needed, then estimates retained candidates again while rebuilding the truncated result. Replace the second estimation with a per-item token-cost list built during the first pass. This preserves layer priority and item ordering.

## Follow-Up Candidates

- Cross-turn retrieval cache design.
- Graph RAG traversal benchmark and limit tuning.
- Orchestration context-injection deduplication.
- Agent fan-in cancellation and throughput profiling.
```

- [ ] **Step 5: Fill OpenSpec tasks**

Write `openspec/changes/retrieval-orchestration-efficiency-audit/tasks.md` with this content:

```markdown
# Tasks

- [ ] Create code-grounded flow map in design.md.
- [ ] Add retrieval aggregation behavior tests.
- [ ] Pre-size retrieval aggregation containers.
- [ ] Add context truncation behavior tests.
- [ ] Avoid duplicate token estimation in `knowledge.TruncateResult`.
- [ ] Run focused package tests.
- [ ] Run `go build ./...`.
- [ ] Run `go test ./...`.
- [ ] Verify OpenSpec change.
- [ ] Sync specs.
- [ ] Archive change.
```

- [ ] **Step 6: Commit OpenSpec artifacts**

Run:

```bash
git add openspec/changes/retrieval-orchestration-efficiency-audit
git -c commit.gpgsign=false commit -m "spec: add retrieval orchestration efficiency audit"
```

Expected:

```text
[feature/retrieval-orchestration-efficiency ...] spec: add retrieval orchestration efficiency audit
```

## Task 2: Add Retrieval Aggregation Tests

**Files:**
- Modify: `internal/retrieval/coordinator_test.go`

- [ ] **Step 1: Add behavior-preserving test for retrieval output**

Append this test to `internal/retrieval/coordinator_test.go`:

```go
func TestRetrievalCoordinator_Retrieve_PreallocationBehavior(t *testing.T) {
	agent1 := &mockAgent{
		name: "agent-1",
		findings: []Finding{
			{
				Key:     "shared",
				Content: "authoritative",
				Score:   0.1,
				Source:  "knowledge",
				Layer:   knowledge.LayerUserKnowledge,
				Agent:   "agent-1",
			},
			{
				Key:     "unique-a",
				Content: "unique high score",
				Score:   0.9,
				Layer:   knowledge.LayerUserKnowledge,
				Agent:   "agent-1",
			},
		},
	}
	agent2 := &mockAgent{
		name: "agent-2",
		findings: []Finding{
			{
				Key:     "shared",
				Content: "less authoritative",
				Score:   9.0,
				Source:  "conversation_analysis",
				Layer:   knowledge.LayerUserKnowledge,
				Agent:   "agent-2",
			},
			{
				Key:     "unique-b",
				Content: "unique medium score",
				Score:   0.5,
				Layer:   knowledge.LayerAgentLearnings,
				Agent:   "agent-2",
			},
		},
	}

	coord := NewRetrievalCoordinator([]RetrievalAgent{agent1, agent2}, zap.NewNop().Sugar())
	got, err := coord.Retrieve(context.Background(), "query", 0)
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("want 3 findings after merge, got %d: %#v", len(got), got)
	}
	for _, f := range got {
		if f.Key == "shared" && f.Agent != "agent-1" {
			t.Fatalf("shared finding should keep higher authority agent-1, got %q", f.Agent)
		}
	}
	if got[0].Score < got[1].Score || got[1].Score < got[2].Score {
		t.Fatalf("findings should remain sorted by score descending: %#v", got)
	}
}
```

- [ ] **Step 2: Run focused retrieval test and verify it passes before implementation**

Run:

```bash
go test ./internal/retrieval -run TestRetrievalCoordinator_Retrieve_PreallocationBehavior -count=1
```

Expected:

```text
ok  	github.com/langoai/lango/internal/retrieval
```

- [ ] **Step 3: Add retrieval benchmark for aggregation path**

Append this benchmark to `internal/retrieval/coordinator_test.go`:

```go
func BenchmarkRetrievalCoordinator_Retrieve_Aggregation(b *testing.B) {
	findings := make([]Finding, 0, 200)
	for i := 0; i < 200; i++ {
		findings = append(findings, Finding{
			Key:     fmt.Sprintf("key-%03d", i),
			Content: "retrieval content for aggregation benchmark",
			Score:   float64(i),
			Layer:   knowledge.LayerUserKnowledge,
			Agent:   "bench-agent",
		})
	}

	agents := []RetrievalAgent{
		&mockAgent{name: "a", findings: findings[:100]},
		&mockAgent{name: "b", findings: findings[100:]},
	}
	coord := NewRetrievalCoordinator(agents, zap.NewNop().Sugar())

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		got, err := coord.Retrieve(context.Background(), "query", 0)
		if err != nil {
			b.Fatalf("Retrieve: %v", err)
		}
		if len(got) != 200 {
			b.Fatalf("want 200 findings, got %d", len(got))
		}
	}
}
```

- [ ] **Step 4: Run benchmark baseline**

Run:

```bash
go test ./internal/retrieval -bench BenchmarkRetrievalCoordinator_Retrieve_Aggregation -benchmem -run '^$'
```

Expected:

```text
BenchmarkRetrievalCoordinator_Retrieve_Aggregation
PASS
```

## Task 3: Pre-Size Retrieval Aggregation Containers

**Files:**
- Modify: `internal/retrieval/coordinator.go`
- Modify: `internal/retrieval/coordinator_test.go`

- [ ] **Step 1: Update `Retrieve` aggregation allocation**

In `internal/retrieval/coordinator.go`, replace:

```go
	var allFindings []Finding
	for _, r := range results {
		allFindings = append(allFindings, r...)
	}
```

with:

```go
	totalFindings := 0
	for _, r := range results {
		totalFindings += len(r)
	}

	allFindings := make([]Finding, 0, totalFindings)
	for _, r := range results {
		allFindings = append(allFindings, r...)
	}
```

- [ ] **Step 2: Run retrieval tests**

Run:

```bash
go test ./internal/retrieval -count=1
```

Expected:

```text
ok  	github.com/langoai/lango/internal/retrieval
```

- [ ] **Step 3: Run benchmark after implementation**

Run:

```bash
go test ./internal/retrieval -bench BenchmarkRetrievalCoordinator_Retrieve_Aggregation -benchmem -run '^$'
```

Expected:

```text
BenchmarkRetrievalCoordinator_Retrieve_Aggregation
PASS
```

- [ ] **Step 4: Commit retrieval aggregation quick win**

Run:

```bash
git add internal/retrieval/coordinator.go internal/retrieval/coordinator_test.go
git -c commit.gpgsign=false commit -m "perf: preallocate retrieval aggregation"
```

Expected:

```text
[feature/retrieval-orchestration-efficiency ...] perf: preallocate retrieval aggregation
```

## Task 4: Add Context Truncation Behavior Tests

**Files:**
- Modify: `internal/knowledge/retriever_test.go`

- [ ] **Step 1: Add test that protects layer priority and budget behavior**

Append this test to `internal/knowledge/retriever_test.go`:

```go
func TestTruncateResult_PreservesLayerPriorityAndOrder(t *testing.T) {
	result := &RetrievalResult{
		Items: map[ContextLayer][]ContextItem{
			LayerUserKnowledge: {
				{Key: "knowledge-1", Content: "short knowledge item"},
			},
			LayerAgentLearnings: {
				{Key: "learning-1", Content: strings.Repeat("learning ", 200)},
			},
			LayerRuntimeContext: {
				{Key: "runtime-1", Content: "runtime context"},
			},
		},
		TotalItems: 3,
	}

	got := TruncateResult(result, 20)
	if got.TotalItems == 0 {
		t.Fatal("expected priority layers to keep at least one item")
	}

	runtimeItems := got.Items[LayerRuntimeContext]
	if len(runtimeItems) != 1 || runtimeItems[0].Key != "runtime-1" {
		t.Fatalf("runtime context should be retained first, got %#v", got.Items)
	}
	if _, ok := got.Items[LayerAgentLearnings]; ok {
		t.Fatalf("large learning item should not fit in small budget: %#v", got.Items)
	}
}
```

- [ ] **Step 2: Run focused knowledge truncation tests before implementation**

Run:

```bash
go test ./internal/knowledge -run 'TestTruncateResult|TestTruncateResult_PreservesLayerPriorityAndOrder' -count=1
```

Expected:

```text
ok  	github.com/langoai/lango/internal/knowledge
```

- [ ] **Step 3: Add truncation benchmark**

Append this benchmark to `internal/knowledge/retriever_test.go`:

```go
func BenchmarkTruncateResult_ManyItems(b *testing.B) {
	items := make([]ContextItem, 0, 500)
	for i := 0; i < 500; i++ {
		items = append(items, ContextItem{
			Key:     fmt.Sprintf("item-%03d", i),
			Content: strings.Repeat("token ", 40),
		})
	}
	result := &RetrievalResult{
		Items:      map[ContextLayer][]ContextItem{LayerUserKnowledge: items},
		TotalItems: len(items),
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		got := TruncateResult(result, 1000)
		if got.TotalItems == 0 {
			b.Fatal("expected truncated result to keep some items")
		}
	}
}
```

- [ ] **Step 4: Add missing import if needed**

If `fmt` is not already imported in `internal/knowledge/retriever_test.go`, update the import block to include it:

```go
import (
	"context"
	"fmt"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/langoai/lango/internal/ent/enttest"
	entlearning "github.com/langoai/lango/internal/ent/learning"
	_ "github.com/mattn/go-sqlite3"
)
```

- [ ] **Step 5: Run benchmark baseline**

Run:

```bash
go test ./internal/knowledge -bench BenchmarkTruncateResult_ManyItems -benchmem -run '^$'
```

Expected:

```text
BenchmarkTruncateResult_ManyItems
PASS
```

## Task 5: Avoid Duplicate Token Estimation in Context Truncation

**Files:**
- Modify: `internal/knowledge/retriever.go`
- Modify: `internal/knowledge/retriever_test.go`

- [ ] **Step 1: Add local helper type near `TruncateResult`**

In `internal/knowledge/retriever.go`, add this type above `TruncateResult`:

```go
type contextItemWithTokens struct {
	item   ContextItem
	tokens int
}
```

- [ ] **Step 2: Replace `TruncateResult` with single-estimation implementation**

Replace the full `TruncateResult` function in `internal/knowledge/retriever.go` with:

```go
func TruncateResult(result *RetrievalResult, budgetTokens int) *RetrievalResult {
	if result == nil || budgetTokens == 0 || result.TotalItems == 0 {
		return result
	}

	itemsWithTokens := make(map[ContextLayer][]contextItemWithTokens, len(result.Items))
	totalTokens := 0
	for layer, items := range result.Items {
		layerItems := make([]contextItemWithTokens, 0, len(items))
		for _, item := range items {
			tokens := types.EstimateTokens(item.Content)
			totalTokens += tokens
			layerItems = append(layerItems, contextItemWithTokens{
				item:   item,
				tokens: tokens,
			})
		}
		itemsWithTokens[layer] = layerItems
	}

	if totalTokens <= budgetTokens {
		return result
	}

	truncated := &RetrievalResult{
		Items: make(map[ContextLayer][]ContextItem),
	}
	remaining := budgetTokens

	priorityOrder := []ContextLayer{
		LayerRuntimeContext, LayerToolRegistry, LayerUserKnowledge,
		LayerSkillPatterns, LayerExternalKnowledge, LayerAgentLearnings,
		LayerPendingInquiries,
	}

	for _, layer := range priorityOrder {
		items, ok := itemsWithTokens[layer]
		if !ok || len(items) == 0 {
			continue
		}

		kept := make([]ContextItem, 0, len(items))
		for _, candidate := range items {
			if remaining-candidate.tokens < 0 {
				break
			}
			kept = append(kept, candidate.item)
			remaining -= candidate.tokens
		}

		if len(kept) > 0 {
			truncated.Items[layer] = kept
			truncated.TotalItems += len(kept)
		}
	}

	return truncated
}
```

- [ ] **Step 3: Run focused knowledge tests**

Run:

```bash
go test ./internal/knowledge -run 'TestTruncateResult|TestTruncateResult_PreservesLayerPriorityAndOrder' -count=1
```

Expected:

```text
ok  	github.com/langoai/lango/internal/knowledge
```

- [ ] **Step 4: Run truncation benchmark after implementation**

Run:

```bash
go test ./internal/knowledge -bench BenchmarkTruncateResult_ManyItems -benchmem -run '^$'
```

Expected:

```text
BenchmarkTruncateResult_ManyItems
PASS
```

- [ ] **Step 5: Commit context truncation quick win**

Run:

```bash
git add internal/knowledge/retriever.go internal/knowledge/retriever_test.go
git -c commit.gpgsign=false commit -m "perf: avoid duplicate context token estimates"
```

Expected:

```text
[feature/retrieval-orchestration-efficiency ...] perf: avoid duplicate context token estimates
```

## Task 6: Verify, Sync, Archive, and Document Outcome

**Files:**
- Modify: `openspec/specs/retrieval-orchestration-efficiency/spec.md`
- Create: `openspec/changes/archive/2026-04-27-retrieval-orchestration-efficiency-audit/**`
- Modify only if user-facing behavior changed: public docs under `docs/`

- [ ] **Step 1: Run focused package tests**

Run:

```bash
go test ./internal/retrieval ./internal/knowledge -count=1
```

Expected:

```text
ok  	github.com/langoai/lango/internal/retrieval
ok  	github.com/langoai/lango/internal/knowledge
```

- [ ] **Step 2: Run full Go verification**

Run:

```bash
go build ./...
go test ./...
```

Expected:

```text
go build ./... exits 0
go test ./... exits 0
```

- [ ] **Step 3: Verify OpenSpec change**

Run through the repository OpenSpec workflow:

```text
/opsx:verify retrieval-orchestration-efficiency-audit
```

Expected:

```text
retrieval-orchestration-efficiency-audit verified
```

- [ ] **Step 4: Sync the delta spec into main specs**

Run:

```text
/opsx:sync retrieval-orchestration-efficiency-audit
```

Expected:

```text
openspec/specs/retrieval-orchestration-efficiency/spec.md updated
```

- [ ] **Step 5: Archive the OpenSpec change**

Run:

```text
/opsx:archive retrieval-orchestration-efficiency-audit
```

Expected:

```text
openspec/changes/archive/2026-04-27-retrieval-orchestration-efficiency-audit/ created
openspec/changes/retrieval-orchestration-efficiency-audit/ removed
```

- [ ] **Step 6: Confirm public docs are not needed**

Inspect changed behavior:

```text
The changes only affect internal allocation/token-estimation efficiency.
No CLI, TUI, README, or public docs update is required.
```

If implementation changed any user-facing command output, update the affected public docs after auditing the CLI/TUI wiring. Otherwise, leave public docs untouched.

- [ ] **Step 7: Commit OpenSpec sync/archive outcome**

Run:

```bash
git add openspec/specs/retrieval-orchestration-efficiency/spec.md openspec/changes/archive/2026-04-27-retrieval-orchestration-efficiency-audit
git -c commit.gpgsign=false commit -m "spec: archive retrieval orchestration efficiency audit"
```

Expected:

```text
[feature/retrieval-orchestration-efficiency ...] spec: archive retrieval orchestration efficiency audit
```

## Self-Review Checklist

- Spec coverage: The plan creates OpenSpec artifacts, maps the retrieval/orchestration flow, implements two low-risk quick wins, verifies behavior, and archives the change.
- Scope control: Ranking, Graph RAG semantics, P2P semantics, runtime redesign, and public docs expansion remain out of scope.
- Type consistency: New helper type uses existing `ContextItem` and `ContextLayer` names from `internal/knowledge`; retrieval tests use existing `Finding`, `RetrievalAgent`, and `mockAgent`.
- Verification: Focused tests, benchmarks, full `go build ./...`, full `go test ./...`, OpenSpec verify/sync/archive are included.
