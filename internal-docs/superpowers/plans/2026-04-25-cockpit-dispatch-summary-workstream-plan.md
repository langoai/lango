# Cockpit Dispatch Summary Workstream Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the cockpit dead-letter summary strip with top 5 latest dispatch references.

**Architecture:** Keep the current dead-letters page and the landed multi-line summary strip intact, then extend that strip additively with a compact latest-dispatch line. The workstream reuses the same backlog rows already loaded by the cockpit page, performs latest-dispatch aggregation in page-local code, and preserves the existing reload-aligned recomputation behavior.

**Tech Stack:** Go, Bubble Tea cockpit UI, `internal/cli/cockpit/pages`, existing dead-letter cockpit page state, Zensical docs, OpenSpec

---

## File Map

### Worker A: Cockpit Code / Tests

- Modify: `internal/cli/cockpit/pages/deadletters.go`
  - Extend page-local summary aggregation and summary-strip rendering with top dispatch references.
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`
  - Cover dispatch aggregation and strip rendering after reload paths.

### Worker B: Docs / OpenSpec / README

- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe richer cockpit summary strip behavior with top dispatch references.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Narrow the remaining richer-cockpit-summary follow-on work.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the public docs requirements for dispatch-rich cockpit summaries.
- Modify: `README.md` only if the cockpit dead-letter page description needs a visible note.
- Create: `openspec/changes/archive/2026-04-25-cockpit-dispatch-summary-workstream/**`
  - Proposal, design, tasks, and delta specs.

## Task Breakdown

### Task 1: Extend the Cockpit Summary Strip with Top Dispatch References

**Owner:** Worker A

**Files:**
- Modify: `internal/cli/cockpit/pages/deadletters.go`
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`

- [ ] **Step 1: Write or extend failing cockpit tests**

Add or extend coverage for:

- top 5 latest dispatch reference aggregation
- additive extension of the existing page-top summary strip
- compact rendering of top dispatch references
- recompute on:
  - initial load
  - filter apply
  - `Ctrl+R` reset
  - retry-success refresh

- [ ] **Step 2: Run the focused cockpit tests and verify they fail**

Run:

```bash
go test ./internal/cli/cockpit/pages -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the richer summary-strip extension**

Extend the existing summary strip so that it:

- keeps:
  - total dead letters
  - retryable count
  - by adjudication
  - by latest family
  - top 5 latest dead-letter reasons
  - top 5 latest manual replay actors
- adds:
  - top 5 latest dispatch references

Implementation rules:

- aggregate from `LatestDispatchReference`
- keep the strip compact
- keep the existing placement
- recompute whenever backlog rows are reloaded
- do not add a new backend summary bridge
- do not add a new pane or page

- [ ] **Step 4: Re-run the focused cockpit tests and verify they pass**

Run:

```bash
go test ./internal/cli/cockpit/pages -count=1
```

Expected:

```text
ok
```

### Task 2: Truth-Align Docs / OpenSpec

**Owner:** Worker B

**Files:**
- Modify: verified docs/OpenSpec files
- Create: `openspec/changes/archive/2026-04-25-cockpit-dispatch-summary-workstream/**`

- [ ] **Step 1: Audit the final dispatch-rich cockpit summary strip before writing docs**

Verify:

- placement on the dead-letters page
- compact multi-line strip shape
- top dispatch fields actually shown
- reload/recompute semantics

- [ ] **Step 2: Update public docs**

Document:

- richer dead-letters page top summary strip
- existing global overview chips
- existing top reasons line
- existing top actors line
- added top latest dispatch references
- shared backlog-reload-aligned refresh behavior

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed dispatch-rich cockpit summary slice.

- [ ] **Step 4: Archive the completed workstream**

Create:

- `openspec/changes/archive/2026-04-25-cockpit-dispatch-summary-workstream/proposal.md`
- `openspec/changes/archive/2026-04-25-cockpit-dispatch-summary-workstream/design.md`
- `openspec/changes/archive/2026-04-25-cockpit-dispatch-summary-workstream/tasks.md`
- `openspec/changes/archive/2026-04-25-cockpit-dispatch-summary-workstream/specs/docs-only/spec.md`

### Task 3: Final Verification and Integration

**Owner:** Main agent

- [ ] **Step 1: Review Worker A changes**

Check:

- summary aggregation correctness
- compact strip density
- no regression in filter/table/detail behavior
- recompute after reload paths

- [ ] **Step 2: Review Worker B changes**

Check:

- docs match the landed cockpit summary behavior
- README additions only describe actual surfaced behavior
- OpenSpec language matches the landed implementation

- [ ] **Step 3: Run full verification**

Run:

```bash
go build ./...
go test ./...
.venv/bin/zensical build
```

Expected:

```text
ok
Build finished
```
