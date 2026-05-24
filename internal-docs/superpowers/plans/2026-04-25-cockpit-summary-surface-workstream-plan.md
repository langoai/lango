# Cockpit Summary Surface Workstream Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add the first cockpit-native dead-letter summary surface as a compact top summary strip on the dead-letters page.

**Architecture:** Keep the current cockpit dead-letter page layout intact and layer a page-top summary strip over the already-loaded backlog rows. The workstream computes a page-local summary from the same backlog result that drives the table/detail surfaces, reusing existing reload semantics rather than adding a new backend summary path or a separate cockpit page.

**Tech Stack:** Go, Bubble Tea cockpit UI, `internal/cli/cockpit/pages`, existing dead-letter cockpit page state, Zensical docs, OpenSpec

---

## File Map

### Worker A: Cockpit Code / Tests

- Modify: `internal/cli/cockpit/pages/deadletters.go`
  - Add page-local summary aggregation and top strip rendering.
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`
  - Cover summary-strip rendering and recomputation after reload paths.

### Worker B: Docs / OpenSpec / README

- Modify: `docs/architecture/dead-letter-browsing-status-observation.md`
  - Describe the cockpit summary strip as landed work.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Narrow the remaining cockpit-summary follow-on work.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync the public docs requirements for cockpit summary surfaces.
- Modify: `README.md` only if the cockpit dead-letter page description needs a visible summary-surface note.
- Create: `openspec/changes/archive/2026-04-25-cockpit-summary-surface-workstream/**`
  - Proposal, design, tasks, and delta specs.

## Task Breakdown

### Task 1: Add the Cockpit Dead-Letter Summary Strip

**Owner:** Worker A

**Files:**
- Modify: `internal/cli/cockpit/pages/deadletters.go`
- Modify: `internal/cli/cockpit/pages/deadletters_test.go`

- [ ] **Step 1: Write or extend failing cockpit tests**

Add or extend coverage for:

- a page-top summary strip on the dead-letters page
- summary aggregation from currently loaded backlog rows
- summary fields:
  - total dead letters
  - retryable count
  - by adjudication
  - by latest family
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

- [ ] **Step 3: Implement the summary strip**

Add:

- page-local summary aggregation helper
- compact single-line summary strip

Implementation rules:

- reuse the existing backlog rows already loaded by the page
- keep the first slice to:
  - total dead letters
  - retryable count
  - by adjudication
  - by latest family
- recompute whenever the backlog rows are reloaded
- do not add a new backend summary bridge
- do not add a new page or pane

- [ ] **Step 4: Re-run the focused cockpit tests and verify they pass**

Run:

```bash
go test ./internal/cli/cockpit/pages -count=1
```

Expected:

```text
ok
```

### Task 2: Truth-Align Docs / README / OpenSpec

**Owner:** Worker B

**Files:**
- Modify: verified docs/OpenSpec files
- Create: `openspec/changes/archive/2026-04-25-cockpit-summary-surface-workstream/**`

- [ ] **Step 1: Audit the final cockpit summary strip before writing docs**

Verify:

- placement on the dead-letters page
- compact single-line strip shape
- fields actually shown
- reload/recompute semantics

- [ ] **Step 2: Update public docs**

Document:

- dead-letters page top summary strip
- global overview fields
- backlog-reload-aligned refresh behavior

- [ ] **Step 3: Sync main OpenSpec requirements**

Update:

- `openspec/specs/docs-only/spec.md`

to reflect the landed cockpit summary slice.

- [ ] **Step 4: Archive the completed workstream**

Create:

- `openspec/changes/archive/2026-04-25-cockpit-summary-surface-workstream/proposal.md`
- `openspec/changes/archive/2026-04-25-cockpit-summary-surface-workstream/design.md`
- `openspec/changes/archive/2026-04-25-cockpit-summary-surface-workstream/tasks.md`
- `openspec/changes/archive/2026-04-25-cockpit-summary-surface-workstream/specs/docs-only/spec.md`

### Task 3: Final Verification and Integration

**Owner:** Main agent

- [ ] **Step 1: Review Worker A changes**

Check:

- summary aggregation correctness
- strip placement and density
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
