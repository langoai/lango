# Actor Family Summary Workstream Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add actor-family grouped summary parity to the dead-letter CLI summary and cockpit summary strip.

**Architecture:** Add one shared actor-family classifier for latest manual replay actor strings, then extend the existing CLI and cockpit summary aggregations additively. The workstream preserves existing raw top latest manual replay actors and adds grouped actor-family buckets without changing backend read models, retry behavior, filters, or summary commands.

**Tech Stack:** Go, Cobra status CLI, Bubble Tea cockpit UI, `internal/cli/status`, `internal/cli/cockpit/pages`, `internal/postadjudicationstatus`, Zensical docs, OpenSpec

---

## File Map

### Shared Helper

- Create: `internal/postadjudicationstatus/actor_family.go`
  - Provides `ClassifyManualReplayActorFamily(actor string) string`.
- Create: `internal/postadjudicationstatus/actor_family_test.go`
  - Covers case-insensitive heuristic matching and `unknown` fallback.

### Worker A: CLI Code / Tests

- Modify: `internal/cli/status/status.go`
  - Add `by_actor_family` to the summary result and aggregate buckets from latest manual replay actors.
- Modify: `internal/cli/status/render.go`
  - Add a `By actor family` table section.
- Modify: `internal/cli/status/status_test.go`
  - Cover JSON/table summary output and aggregation behavior.

### Worker B: Cockpit Code / Tests

- Modify: `internal/cli/cockpit/pages/deadletters.go`
  - Add actor-family buckets to page-local summary aggregation and render a compact `actor families:` strip line.
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`
  - Cover actor-family aggregation and strip rendering.

### Worker C: Docs / OpenSpec / README

- Modify: `docs/cli/status.md`
  - Document `by_actor_family` in the summary command.
- Modify: `docs/cli/index.md` if command summary copy changes.
- Modify: `README.md` if the status command inventory changes.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe grouped actor-family summaries for CLI and cockpit.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Update landed/remaining work wording.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync public docs requirements.
- Create: `openspec/changes/archive/2026-04-26-actor-family-summary-workstream/**`
  - Proposal, design, tasks, and delta spec.

## Task Breakdown

### Task 1: Add the Shared Actor-Family Classifier

**Owner:** Worker A

**Files:**
- Create: `internal/postadjudicationstatus/actor_family.go`
- Create: `internal/postadjudicationstatus/actor_family_test.go`

- [ ] **Step 1: Write focused classifier tests**

Create `internal/postadjudicationstatus/actor_family_test.go` with table tests for:

- operator identities mapping to `operator`
- system / runtime / auto identities mapping to `system`
- service / bridge / integration identities mapping to `service`
- empty or unmatched strings mapping to `unknown`
- case-insensitive matching

Expected test shape:

```go
func TestClassifyManualReplayActorFamily(t *testing.T) {
	tests := []struct {
		name  string
		actor string
		want  string
	}{
		{name: "operator", actor: "operator:alice", want: "operator"},
		{name: "system", actor: "system:auto-retry", want: "system"},
		{name: "service", actor: "service:bridge", want: "service"},
		{name: "unknown empty", actor: "", want: "unknown"},
		{name: "unknown unmatched", actor: "alice", want: "unknown"},
		{name: "case insensitive", actor: "OPERATOR:BOB", want: "operator"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, ClassifyManualReplayActorFamily(tt.actor))
		})
	}
}
```

- [ ] **Step 2: Run classifier tests and verify they fail**

Run:

```bash
go test ./internal/postadjudicationstatus -run TestClassifyManualReplayActorFamily -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the classifier**

Create `internal/postadjudicationstatus/actor_family.go` with:

```go
package postadjudicationstatus

import "strings"

const (
	ManualReplayActorFamilyOperator = "operator"
	ManualReplayActorFamilySystem   = "system"
	ManualReplayActorFamilyService  = "service"
	ManualReplayActorFamilyUnknown  = "unknown"
)

func ClassifyManualReplayActorFamily(actor string) string {
	normalized := strings.ToLower(strings.TrimSpace(actor))
	if normalized == "" {
		return ManualReplayActorFamilyUnknown
	}
	if hasAnyPrefix(normalized, "operator:", "user:", "human:") {
		return ManualReplayActorFamilyOperator
	}
	if hasAnyPrefix(normalized, "system:", "runtime:", "auto:", "worker:") {
		return ManualReplayActorFamilySystem
	}
	if hasAnyPrefix(normalized, "service:", "bridge:", "integration:", "bot:") {
		return ManualReplayActorFamilyService
	}
	return ManualReplayActorFamilyUnknown
}

func hasAnyPrefix(value string, prefixes ...string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}
```

Adjust the exact prefix set only if tests or existing manual replay actor fixtures show a clearer local convention.

- [ ] **Step 4: Run classifier tests and verify they pass**

Run:

```bash
go test ./internal/postadjudicationstatus -run TestClassifyManualReplayActorFamily -count=1
```

Expected:

```text
ok
```

### Task 2: Extend the CLI Summary with `by_actor_family`

**Owner:** Worker A

**Files:**
- Modify: `internal/cli/status/status.go`
- Modify: `internal/cli/status/render.go`
- Modify: `internal/cli/status/status_test.go`

- [ ] **Step 1: Write or extend failing CLI summary tests**

Add coverage that verifies:

- `aggregateDeadLetterSummary` includes `ByActorFamily`
- JSON output contains `by_actor_family`
- table output contains `By actor family`
- raw `top_latest_manual_replay_actors` remains present

Use entries with latest manual replay actors that map to at least:

- `operator`
- `system`
- `service`
- `unknown`

- [ ] **Step 2: Run focused CLI tests and verify they fail**

Run:

```bash
go test ./internal/cli/status -run 'TestDeadLetterSummary|TestAggregateDeadLetterSummary' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Extend the CLI summary result type and aggregation**

In `internal/cli/status/status.go`:

- add `ByActorFamily []deadLetterSummaryBucket 'json:"by_actor_family"'` to `deadLetterSummaryResult`
- add `actorFamilyCounts := map[string]int{}` inside `aggregateDeadLetterSummary`
- for every row, call `postadjudicationstatus.ClassifyManualReplayActorFamily(entry.LatestManualReplayActor)` and increment that family
- return ordered buckets with preferred order:

```go
[]string{
	postadjudicationstatus.ManualReplayActorFamilyOperator,
	postadjudicationstatus.ManualReplayActorFamilySystem,
	postadjudicationstatus.ManualReplayActorFamilyService,
	postadjudicationstatus.ManualReplayActorFamilyUnknown,
}
```

- [ ] **Step 4: Extend table rendering**

In `internal/cli/status/render.go`, add a `By actor family` section in `renderDeadLetterSummaryTable`, near the existing `By reason family` / top actor sections:

```go
b.WriteString(sectionHeader("By actor family"))
b.WriteString(renderSummaryBuckets(summary.ByActorFamily))
```

Preserve the existing top latest manual replay actors section.

- [ ] **Step 5: Run focused CLI tests and verify they pass**

Run:

```bash
go test ./internal/cli/status -run 'TestDeadLetterSummary|TestAggregateDeadLetterSummary' -count=1
```

Expected:

```text
ok
```

### Task 3: Extend the Cockpit Summary Strip with Actor Families

**Owner:** Worker B

**Files:**
- Modify: `internal/cli/cockpit/pages/deadletters.go`
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`

- [ ] **Step 1: Write or extend failing cockpit tests**

Add coverage that verifies:

- `summarizeDeadLetters` aggregates actor-family buckets from latest manual replay actors
- the cockpit view renders a compact `actor families:` line
- the existing lines remain visible:
  - global overview
  - `reasons:`
  - `reason families:`
  - `actors:`
  - `dispatch:`
- the family line uses the same family labels as the shared classifier

- [ ] **Step 2: Run focused cockpit tests and verify they fail**

Run:

```bash
go test ./internal/cli/cockpit/pages -run 'TestDeadLetters.*Summary|TestSummarizeDeadLetters' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Extend cockpit summary aggregation**

In `internal/cli/cockpit/pages/deadletters.go`:

- add an actor-family summary item type if no existing shared bucket type fits the page style
- add `actorFamilyCounts := make(map[string]int)` in `summarizeDeadLetters`
- call `postadjudicationstatus.ClassifyManualReplayActorFamily(item.LatestManualReplayActor)` for each row
- order buckets by:

```go
[]string{"operator", "system", "service", "unknown"}
```

Prefer using the exported constants from `postadjudicationstatus` instead of hard-coded strings.

- [ ] **Step 4: Extend strip rendering**

In `renderSummaryStrip`, render a compact line after `actors:` and before `dispatch:`:

```text
actor families: operator(4), system(2), unknown(1)
```

Keep the line hidden only when there are no backlog rows. With at least one row, `unknown` may appear because the classifier has a fallback.

- [ ] **Step 5: Run focused cockpit tests and verify they pass**

Run:

```bash
go test ./internal/cli/cockpit/pages -run 'TestDeadLetters.*Summary|TestSummarizeDeadLetters' -count=1
```

Expected:

```text
ok
```

### Task 4: Truth-Align Docs / OpenSpec

**Owner:** Worker C

**Files:**
- Modify: `docs/cli/status.md`
- Modify: `docs/cli/index.md` if command summary copy changes
- Modify: `README.md` if command inventory copy changes
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-26-actor-family-summary-workstream/**`

- [ ] **Step 1: Audit final CLI and cockpit behavior before writing docs**

Verify from code:

- CLI summary has `by_actor_family`
- CLI table has a `By actor family` section
- cockpit strip has an `actor families:` line
- raw top latest manual replay actors remain documented and rendered
- classifier is heuristic and local to operator summaries

- [ ] **Step 2: Update public docs**

Update public docs to describe:

- the `by_actor_family` CLI summary field
- the cockpit `actor families:` strip line
- the initial actor-family taxonomy:
  - `operator`
  - `system`
  - `service`
  - `unknown`
- the fact that raw top latest manual replay actors remain available

- [ ] **Step 3: Sync main OpenSpec requirements**

Update `openspec/specs/docs-only/spec.md` so the docs-only requirements mention:

- grouped actor-family summary for `lango status dead-letter-summary`
- grouped actor-family line for cockpit dead-letter summary strip
- remaining grouped work as dispatch families, configurable taxonomy, and trend/time-window summaries

- [ ] **Step 4: Archive the completed workstream**

Create:

- `openspec/changes/archive/2026-04-26-actor-family-summary-workstream/proposal.md`
- `openspec/changes/archive/2026-04-26-actor-family-summary-workstream/design.md`
- `openspec/changes/archive/2026-04-26-actor-family-summary-workstream/tasks.md`
- `openspec/changes/archive/2026-04-26-actor-family-summary-workstream/specs/docs-only/spec.md`

The archive should describe the landed docs/spec sync and the implementation scope.

### Task 5: Final Verification and Integration

**Owner:** Main agent

- [ ] **Step 1: Review shared classifier changes**

Check:

- prefix ordering does not misclassify broad service/runtime values before operator-specific prefixes
- empty/unmatched strings return `unknown`
- constants are used by both summary surfaces where practical

- [ ] **Step 2: Review CLI changes**

Check:

- JSON field name is exactly `by_actor_family`
- table section is readable and does not remove existing raw top actors
- bucket ordering matches the preferred taxonomy order

- [ ] **Step 3: Review cockpit changes**

Check:

- `actor families:` line is compact
- existing `reasons:`, `reason families:`, `actors:`, and `dispatch:` lines remain visible
- summary recomputes from current backlog rows

- [ ] **Step 4: Run full verification**

Run:

```bash
go build ./...
go test ./...
.venv/bin/zensical build
```

Expected:

```text
go build ./... exits 0
go test ./... exits 0
.venv/bin/zensical build exits 0
```

- [ ] **Step 5: Commit the implementation**

Commit message:

```bash
git -c commit.gpgsign=false commit -m "feat: add grouped dead letter actor summaries"
```
