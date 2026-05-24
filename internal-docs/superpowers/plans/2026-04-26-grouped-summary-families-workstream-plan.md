# Grouped Summary Families Workstream Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add reason-family grouped summary parity to the dead-letter CLI summary and cockpit summary strip.

**Architecture:** Add one shared reason-family classifier for latest dead-letter reason strings, then extend the existing CLI and cockpit summary aggregations additively. The workstream preserves existing raw top latest dead-letter reasons and adds grouped reason-family buckets without changing backend read models, retry behavior, filters, or summary commands.

**Tech Stack:** Go, Cobra status CLI, Bubble Tea cockpit UI, `internal/cli/status`, `internal/cli/cockpit/pages`, `internal/postadjudicationstatus`, Zensical docs, OpenSpec

---

## File Map

### Shared Helper

- Create: `internal/postadjudicationstatus/reason_family.go`
  - Provides `ClassifyDeadLetterReasonFamily(reason string) string`.
- Create: `internal/postadjudicationstatus/reason_family_test.go`
  - Covers case-insensitive heuristic matching and `unknown` fallback.

### Worker A: CLI Code / Tests

- Modify: `internal/cli/status/status.go`
  - Add `by_reason_family` to the summary result and aggregate buckets from latest dead-letter reasons.
- Modify: `internal/cli/status/render.go`
  - Add a `By reason family` table section.
- Modify: `internal/cli/status/status_test.go`
  - Cover JSON/table summary output and aggregation behavior.

### Worker B: Cockpit Code / Tests

- Modify: `internal/cli/cockpit/pages/deadletters.go`
  - Add reason-family buckets to page-local summary aggregation and render a compact `reason families:` strip line.
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`
  - Cover reason-family aggregation and strip rendering.

### Worker C: Docs / OpenSpec / README

- Modify: `docs/cli/status.md`
  - Document `by_reason_family` in the summary command.
- Modify: `docs/cli/index.md` if command summary copy changes.
- Modify: `README.md` if the status command inventory changes.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe grouped reason-family summaries for CLI and cockpit.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Update landed/remaining work wording.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync public docs requirements.
- Create: `openspec/changes/archive/2026-04-26-grouped-summary-families-workstream/**`
  - Proposal, design, tasks, and delta spec.

## Task Breakdown

### Task 1: Add the Shared Reason-Family Classifier

**Owner:** Worker A

**Files:**
- Create: `internal/postadjudicationstatus/reason_family.go`
- Create: `internal/postadjudicationstatus/reason_family_test.go`

- [ ] **Step 1: Write focused classifier tests**

Create `internal/postadjudicationstatus/reason_family_test.go` with table tests for:

- retry exhaustion strings mapping to `retry-exhausted`
- policy / gate / denied strings mapping to `policy-blocked`
- receipt / adjudication / transaction invalid strings mapping to `receipt-invalid`
- background / dispatch / worker / task failure strings mapping to `background-failed`
- empty or unmatched strings mapping to `unknown`
- case-insensitive matching

Expected test shape:

```go
func TestClassifyDeadLetterReasonFamily(t *testing.T) {
	tests := []struct {
		name   string
		reason string
		want   string
	}{
		{name: "retry exhausted", reason: "retry attempts exhausted after 5 attempts", want: "retry-exhausted"},
		{name: "policy blocked", reason: "policy gate denied replay", want: "policy-blocked"},
		{name: "invalid receipt", reason: "invalid transaction receipt evidence", want: "receipt-invalid"},
		{name: "background failed", reason: "background dispatch worker failed", want: "background-failed"},
		{name: "unknown empty", reason: "", want: "unknown"},
		{name: "unknown unmatched", reason: "unexpected storage condition", want: "unknown"},
		{name: "case insensitive", reason: "POLICY BLOCKED BY GATE", want: "policy-blocked"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, ClassifyDeadLetterReasonFamily(tt.reason))
		})
	}
}
```

- [ ] **Step 2: Run classifier tests and verify they fail**

Run:

```bash
go test ./internal/postadjudicationstatus -run TestClassifyDeadLetterReasonFamily -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the classifier**

Create `internal/postadjudicationstatus/reason_family.go` with:

```go
package postadjudicationstatus

import "strings"

const (
	DeadLetterReasonFamilyRetryExhausted  = "retry-exhausted"
	DeadLetterReasonFamilyPolicyBlocked   = "policy-blocked"
	DeadLetterReasonFamilyReceiptInvalid  = "receipt-invalid"
	DeadLetterReasonFamilyBackgroundFailed = "background-failed"
	DeadLetterReasonFamilyUnknown         = "unknown"
)

func ClassifyDeadLetterReasonFamily(reason string) string {
	normalized := strings.ToLower(strings.TrimSpace(reason))
	if normalized == "" {
		return DeadLetterReasonFamilyUnknown
	}
	if containsAny(normalized, "retry exhausted", "attempts exhausted", "max retry", "retry limit") {
		return DeadLetterReasonFamilyRetryExhausted
	}
	if containsAny(normalized, "policy", "gate denied", "blocked", "not allowed", "forbidden") {
		return DeadLetterReasonFamilyPolicyBlocked
	}
	if containsAny(normalized, "invalid receipt", "receipt", "adjudication missing", "transaction receipt", "evidence invalid") {
		return DeadLetterReasonFamilyReceiptInvalid
	}
	if containsAny(normalized, "background", "dispatch", "worker", "task failed", "queue") {
		return DeadLetterReasonFamilyBackgroundFailed
	}
	return DeadLetterReasonFamilyUnknown
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
```

Adjust the exact keyword set only if tests or existing dead-letter reason fixtures show a clearer local convention.

- [ ] **Step 4: Run classifier tests and verify they pass**

Run:

```bash
go test ./internal/postadjudicationstatus -run TestClassifyDeadLetterReasonFamily -count=1
```

Expected:

```text
ok
```

### Task 2: Extend the CLI Summary with `by_reason_family`

**Owner:** Worker A

**Files:**
- Modify: `internal/cli/status/status.go`
- Modify: `internal/cli/status/render.go`
- Modify: `internal/cli/status/status_test.go`

- [ ] **Step 1: Write or extend failing CLI summary tests**

Add coverage that verifies:

- `aggregateDeadLetterSummary` includes `ByReasonFamily`
- JSON output contains `by_reason_family`
- table output contains `By reason family`
- raw `top_latest_dead_letter_reasons` remains present

Use entries with latest dead-letter reasons that map to at least:

- `retry-exhausted`
- `policy-blocked`
- `receipt-invalid`
- `background-failed`
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

- add `ByReasonFamily []deadLetterSummaryBucket 'json:"by_reason_family"'` to `deadLetterSummaryResult`
- add `reasonFamilyCounts := map[string]int{}` inside `aggregateDeadLetterSummary`
- for every row, call `postadjudicationstatus.ClassifyDeadLetterReasonFamily(entry.LatestDeadLetterReason)` and increment that family
- return ordered buckets with preferred order:

```go
[]string{
	postadjudicationstatus.DeadLetterReasonFamilyRetryExhausted,
	postadjudicationstatus.DeadLetterReasonFamilyPolicyBlocked,
	postadjudicationstatus.DeadLetterReasonFamilyReceiptInvalid,
	postadjudicationstatus.DeadLetterReasonFamilyBackgroundFailed,
	postadjudicationstatus.DeadLetterReasonFamilyUnknown,
}
```

- [ ] **Step 4: Extend table rendering**

In `internal/cli/status/render.go`, add a `By reason family` section in `renderDeadLetterSummaryTable`, near the existing `By latest family` / top reason sections:

```go
b.WriteString(sectionTitle("By reason family"))
b.WriteString(renderSummaryBuckets(summary.ByReasonFamily))
```

Preserve the existing top latest dead-letter reasons section.

- [ ] **Step 5: Run focused CLI tests and verify they pass**

Run:

```bash
go test ./internal/cli/status -run 'TestDeadLetterSummary|TestAggregateDeadLetterSummary' -count=1
```

Expected:

```text
ok
```

### Task 3: Extend the Cockpit Summary Strip with Reason Families

**Owner:** Worker B

**Files:**
- Modify: `internal/cli/cockpit/pages/deadletters.go`
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`

- [ ] **Step 1: Write or extend failing cockpit tests**

Add coverage that verifies:

- `summarizeDeadLetters` aggregates reason-family buckets from latest dead-letter reasons
- the cockpit view renders a compact `reason families:` line
- the existing lines remain visible:
  - global overview
  - `reasons:`
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

- add a reason-family summary item type if no existing shared bucket type fits the page style
- add `reasonFamilyCounts := make(map[string]int)` in `summarizeDeadLetters`
- call `postadjudicationstatus.ClassifyDeadLetterReasonFamily(item.LatestDeadLetterReason)` for each row
- order buckets by:

```go
[]string{"retry-exhausted", "policy-blocked", "receipt-invalid", "background-failed", "unknown"}
```

Prefer using the exported constants from `postadjudicationstatus` instead of hard-coded strings.

- [ ] **Step 4: Extend strip rendering**

In `renderSummaryStrip`, render a compact line after `reasons:` and before `actors:`:

```text
reason families: policy-blocked(4), retry-exhausted(3), unknown(1)
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
- Create: `openspec/changes/archive/2026-04-26-grouped-summary-families-workstream/**`

- [ ] **Step 1: Audit final CLI and cockpit behavior before writing docs**

Verify from code:

- CLI summary has `by_reason_family`
- CLI table has a `By reason family` section
- cockpit strip has a `reason families:` line
- raw top latest dead-letter reasons remain documented and rendered
- classifier is heuristic and local to operator summaries

- [ ] **Step 2: Update public docs**

Update public docs to describe:

- the `by_reason_family` CLI summary field
- the cockpit `reason families:` strip line
- the initial reason-family taxonomy:
  - `retry-exhausted`
  - `policy-blocked`
  - `receipt-invalid`
  - `background-failed`
  - `unknown`
- the fact that raw top latest dead-letter reasons remain available

- [ ] **Step 3: Sync main OpenSpec requirements**

Update `openspec/specs/docs-only/spec.md` so the docs-only requirements mention:

- grouped reason-family summary for `lango status dead-letter-summary`
- grouped reason-family line for cockpit dead-letter summary strip
- remaining grouped work as actor/dispatch families, configurable taxonomy, and trend/time-window summaries

- [ ] **Step 4: Archive the completed workstream**

Create:

- `openspec/changes/archive/2026-04-26-grouped-summary-families-workstream/proposal.md`
- `openspec/changes/archive/2026-04-26-grouped-summary-families-workstream/design.md`
- `openspec/changes/archive/2026-04-26-grouped-summary-families-workstream/tasks.md`
- `openspec/changes/archive/2026-04-26-grouped-summary-families-workstream/specs/docs-only/spec.md`

The archive should describe the landed docs/spec sync and the implementation scope.

### Task 5: Final Verification and Integration

**Owner:** Main agent

- [ ] **Step 1: Review shared classifier changes**

Check:

- keyword ordering does not misclassify broad receipt/background text before policy/retry text
- empty/unmatched strings return `unknown`
- constants are used by both summary surfaces where practical

- [ ] **Step 2: Review CLI changes**

Check:

- JSON field name is exactly `by_reason_family`
- table section is readable and does not remove existing raw top reasons
- bucket ordering matches the preferred taxonomy order

- [ ] **Step 3: Review cockpit changes**

Check:

- `reason families:` line is compact
- existing `reasons:`, `actors:`, and `dispatch:` lines remain visible
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
git -c commit.gpgsign=false commit -m "feat: add grouped dead letter reason summaries"
```
