# Operator Surface Consolidation Workstream Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Close the remaining dead-letter operator-surface gaps in one workstream by landing grouped dispatch-family summaries, CLI `any_match_family`, richer summary evolution, and richer retry follow-up UX.

**Architecture:** Keep the existing dead-letter CLI and cockpit surfaces intact, then extend them additively from shared operator helpers where duplication would otherwise drift. The workstream stays entirely in the operator layer: `internal/cli/status/*`, `internal/cli/cockpit/*`, and docs/OpenSpec, without redesigning the retry substrate or broader policy runtime.

**Tech Stack:** Go, Cobra status CLI, Bubble Tea cockpit UI, `internal/cli/status`, `internal/cli/cockpit/pages`, `internal/postadjudicationstatus`, Zensical docs, OpenSpec

---

## File Map

### Shared Helpers

- Create or modify focused helper files under `internal/postadjudicationstatus/` only if needed for:
  - dispatch-family classification
  - shared summary bucket ordering
  - shared trend / time-window aggregation helpers

### Worker A: CLI Code / Tests

- Modify: `internal/cli/status/status.go`
  - Extend dead-letter list options with `any_match_family`.
  - Extend CLI summary result/output model for grouped dispatch-family summaries.
  - Add richer retry follow-up result / polling handling.
  - Add summary evolution support for richer top-N and trend / time-window outputs.
- Modify: `internal/cli/status/render.go`
  - Render grouped dispatch-family buckets.
  - Render richer summary evolution sections.
  - Render richer retry follow-up output.
- Modify: `internal/cli/status/status_test.go`
  - Cover `any_match_family`, dispatch-family summary, summary evolution, and retry follow-up behavior.

### Worker B: Cockpit Code / Tests

- Modify: `internal/cli/cockpit/pages/deadletters.go`
  - Extend page-top summary strip with grouped dispatch-family summaries.
  - Add summary evolution rendering for richer top-N and trend / time-window views.
  - Add richer retry follow-up refresh / status interpretation UX.
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`
  - Cover dispatch-family strip output.
  - Cover richer summary evolution rendering.
  - Cover richer retry follow-up UX behavior.

### Worker C: Docs / OpenSpec / README

- Modify: `docs/cli/status.md`
  - Document `--any-match-family`, grouped dispatch-family summaries, richer summary outputs, and richer retry follow-up UX.
- Modify: `docs/cli/index.md`
  - Update command summary copy if the surfaced behavior changes.
- Modify: `README.md`
  - Update command inventory summary only if the status surface description changes materially.
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Truth-align CLI and cockpit dead-letter operator surfaces.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Update landed/remaining operator-surface scope.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync docs-only requirements.
- Create: `openspec/changes/archive/2026-04-26-operator-surface-consolidation-workstream/**`
  - Proposal, design, tasks, and docs-only delta spec.

### Worker D: Optional Shared-Helper / Review Support

- Optional write scope only if needed:
  - focused helper extraction for dispatch-family / summary evolution / retry follow-up support
  - no independent product behavior ownership

## Task Breakdown

### Task 1: Add Shared Dispatch-Family and Summary-Evolution Helpers

**Owner:** Worker A or Worker D

**Files:**
- Create or modify only the minimal helper files actually needed under `internal/postadjudicationstatus/`

- [ ] **Step 1: Write focused helper tests**

Add focused tests for:

- grouped dispatch-family classification from `LatestDispatchReference`
- preferred dispatch-family ordering
- any shared trend / time-window aggregation helpers if introduced

Expected coverage themes:

- known dispatch families map consistently
- unknown / empty values fall back deterministically
- helper output order is stable

- [ ] **Step 2: Run focused helper tests and verify they fail**

Run the narrowest package-level test command that covers the new helper.

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the minimal shared helper**

Implementation rules:

- keep helper scope narrow and operator-surface oriented
- avoid inventing a general analytics subsystem
- expose only the classifier / ordering / aggregation helpers actually reused by both CLI and cockpit

- [ ] **Step 4: Re-run focused helper tests and verify they pass**

Run the same focused helper test command.

Expected:

```text
ok
```

### Task 2: Land CLI `any_match_family` Parity

**Owner:** Worker A

**Files:**
- Modify: `internal/cli/status/status.go`
- Modify: `internal/cli/status/status_test.go`

- [ ] **Step 1: Write or extend failing CLI list tests**

Add coverage that verifies:

- `lango status dead-letters` accepts `--any-match-family`
- supported values are validated
- the selected value is forwarded through the existing dead-letter list bridge
- detail and retry commands are unaffected

- [ ] **Step 2: Run focused CLI tests and verify they fail**

Run:

```bash
go test ./internal/cli/status -run 'TestDeadLetters|TestDeadLetter.*List' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the CLI filter parity**

Implementation rules:

- add `AnyMatchFamily` to the dead-letter list options if not already present in the CLI layer
- validate only the currently supported family values
- reuse the existing list bridge
- do not redesign command structure

- [ ] **Step 4: Re-run focused CLI tests and verify they pass**

Run the same focused test command.

Expected:

```text
ok
```

### Task 3: Land CLI Summary Evolution and Retry Follow-Up UX

**Owner:** Worker A

**Files:**
- Modify: `internal/cli/status/status.go`
- Modify: `internal/cli/status/render.go`
- Modify: `internal/cli/status/status_test.go`

- [ ] **Step 1: Write or extend failing CLI tests**

Add coverage for:

- grouped dispatch-family summary buckets
- richer top-N summary behavior
- trend / time-window summary rendering or payloads
- richer retry follow-up UX after request acceptance
  - polling or follow-up refresh semantics
  - richer structured retry result fields

- [ ] **Step 2: Run focused CLI tests and verify they fail**

Run:

```bash
go test ./internal/cli/status -run 'TestDeadLetterSummary|TestDeadLetterRetry|TestAggregateDeadLetterSummary' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement additive CLI evolution**

Implementation rules:

- keep existing summary commands and retry command structure
- preserve current request flow for retry
- make richer follow-up behavior additive
- keep raw top latest values visible while adding grouped / richer summary sections

- [ ] **Step 4: Re-run focused CLI tests and verify they pass**

Run the same focused test command.

Expected:

```text
ok
```

### Task 4: Land Cockpit Summary Evolution and Retry Follow-Up UX

**Owner:** Worker B

**Files:**
- Modify: `internal/cli/cockpit/pages/deadletters.go`
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`

- [ ] **Step 1: Write or extend failing cockpit tests**

Add coverage for:

- grouped dispatch-family strip output
- richer top-N or trend / time-window summary presentation in the page-top strip
- richer retry follow-up refresh / post-acceptance status behavior
- preservation of current summary-strip and detail-pane semantics

- [ ] **Step 2: Run focused cockpit tests and verify they fail**

Run:

```bash
go test ./internal/cli/cockpit/pages -run 'TestDeadLetters.*Summary|TestDeadLetters.*Retry|TestSummarizeDeadLetters' -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement additive cockpit evolution**

Implementation rules:

- preserve the current dead-letter page layout
- extend the top strip instead of creating a new pane unless a helper view is absolutely needed
- preserve current retry request semantics
- improve follow-up interpretation after request acceptance without changing control-plane behavior

- [ ] **Step 4: Re-run focused cockpit tests and verify they pass**

Run the same focused cockpit test command.

Expected:

```text
ok
```

### Task 5: Truth-Align Docs / OpenSpec

**Owner:** Worker C

**Files:**
- Modify: `docs/cli/status.md`
- Modify: `docs/cli/index.md`
- Modify: `README.md`
- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Create: `openspec/changes/archive/2026-04-26-operator-surface-consolidation-workstream/**`

- [ ] **Step 1: Audit final CLI and cockpit behavior before writing docs**

Verify from code:

- `--any-match-family` is actually wired
- grouped dispatch-family summaries are actually surfaced
- richer summary evolution behavior actually exists
- richer retry follow-up UX actually exists

- [ ] **Step 2: Update public docs**

Document:

- dead-letter CLI filter parity including `--any-match-family`
- grouped dispatch-family summaries
- richer top-N / trend / time-window behavior in the form actually landed
- richer retry follow-up UX in the form actually landed

- [ ] **Step 3: Sync main OpenSpec requirements**

Update `openspec/specs/docs-only/spec.md` so the docs-only requirements reflect the landed operator-surface consolidation slice and narrow the remaining operator backlog accordingly.

- [ ] **Step 4: Archive the completed workstream**

Create:

- `openspec/changes/archive/2026-04-26-operator-surface-consolidation-workstream/proposal.md`
- `openspec/changes/archive/2026-04-26-operator-surface-consolidation-workstream/design.md`
- `openspec/changes/archive/2026-04-26-operator-surface-consolidation-workstream/tasks.md`
- `openspec/changes/archive/2026-04-26-operator-surface-consolidation-workstream/specs/docs-only/spec.md`

### Task 6: Final Verification and Integration

**Owner:** Main agent

- [ ] **Step 1: Review shared helper changes**

Check:

- helper scope stayed narrow
- no accidental general analytics subsystem appeared
- CLI and cockpit reuse the same semantics where intended

- [ ] **Step 2: Review CLI changes**

Check:

- `--any-match-family` is validated and forwarded correctly
- summary evolution is additive
- retry follow-up UX is clearer without changing retry semantics

- [ ] **Step 3: Review cockpit changes**

Check:

- summary strip remains readable
- retry follow-up UX remains aligned with current request-acceptance semantics
- no regressions in selection / reset / refresh behavior

- [ ] **Step 4: Run full verification**

Run:

```bash
go build ./...
go test ./...
.venv/bin/zensical build
openspec validate docs-only --type spec --strict --no-interactive
```

Expected:

```text
go build ./... exits 0
go test ./... exits 0
.venv/bin/zensical build exits 0
openspec validate docs-only --type spec --strict --no-interactive exits 0
```

- [ ] **Step 5: Commit the implementation**

Commit message:

```bash
git -c commit.gpgsign=false commit -m "feat: consolidate operator dead letter surfaces"
```
