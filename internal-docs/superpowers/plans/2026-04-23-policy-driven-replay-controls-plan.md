# Policy-Driven Replay Controls Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce the first policy-driven replay control slice that adds actor- and outcome-aware authorization to `retry_post_adjudication_execution` while keeping the canonical replay gate unchanged.

**Architecture:** Extend the replay service so it resolves the current actor from runtime context, evaluates a simple config-backed allowlist policy, and denies replay when the actor is unresolved or not permitted for the current replay outcome. Keep this gate inside the replay service so canonical replay checks and authorization checks remain in one place. Update docs and OpenSpec to describe the new replay authorization layer.

**Tech Stack:** Go, `internal/postadjudicationreplay`, config-backed policy, existing session / approval context, Zensical docs, OpenSpec

---

## File Map

- Modify: `internal/postadjudicationreplay/types.go`
  - Add replay-policy deny reasons and any policy-related result fields needed.
- Modify: `internal/postadjudicationreplay/service.go`
  - Add actor resolution and outcome-aware policy gate.
- Modify: `internal/postadjudicationreplay/service_test.go`
  - Cover actor-unresolved and replay-not-allowed cases.
- Modify: config typing and defaults under `internal/config/`
  - Add simple replay allowlist fields.
- Modify: `docs/architecture/operator-replay-manual-retry.md`
  - Truth-align the replay page with policy gating.
- Create: `docs/architecture/policy-driven-replay-controls.md`
  - Public architecture/operator doc for the first replay-policy slice.
- Modify: `docs/architecture/index.md`
  - Add the new page to Architecture.
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
  - Mark policy-driven replay controls as landed and push richer policy work down one level.
- Modify: `zensical.toml`
  - Add the new page to Architecture nav.
- Create: `openspec/changes/archive/2026-04-23-policy-driven-replay-controls/**`
  - Proposal, design, tasks, and delta specs.
- Modify: `openspec/specs/project-docs/spec.md`
  - Sync architecture landing requirements.
- Modify: `openspec/specs/docs-only/spec.md`
  - Sync track and page references.
- Modify: `openspec/specs/meta-tools/spec.md`
  - Sync replay authorization contract.

### Task 1: Add Replay-Service Policy Gate

**Files:**
- Modify: `internal/postadjudicationreplay/types.go`
- Modify: `internal/postadjudicationreplay/service.go`
- Modify: `internal/postadjudicationreplay/service_test.go`

- [ ] **Step 1: Write the failing replay service tests**

Add tests covering:

- replay denied when actor cannot be resolved
- replay denied when actor is resolved but not allowed
- replay allowed when actor is permitted for the current outcome
- existing dead-letter / adjudication gates still behave the same

- [ ] **Step 2: Run the replay service tests and verify they fail**

Run:

```bash
go test ./internal/postadjudicationreplay/... -count=1
```

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement actor resolution and policy gate**

Update the replay service so that it:

- resolves actor identity from runtime context
- evaluates:
  - global replay allowlist
  - outcome-specific allowlist
- fails closed when actor cannot be resolved
- fails closed when replay is not permitted

New deny reasons:

- `actor_unresolved`
- `replay_not_allowed`

- [ ] **Step 4: Re-run the replay service tests and verify they pass**

Run:

```bash
go test ./internal/postadjudicationreplay/... -count=1
```

Expected:

```text
ok
```

- [ ] **Step 5: Commit the replay policy slice**

Run:

```bash
git add internal/postadjudicationreplay/types.go internal/postadjudicationreplay/service.go internal/postadjudicationreplay/service_test.go
git -c commit.gpgsign=false commit -m "feat: add replay policy gate"
```

### Task 2: Add Config-Backed Replay Allowlist

**Files:**
- Modify: config typing and defaults under `internal/config/`
- Add or update config tests as needed

- [ ] **Step 1: Write the failing config tests**

Add tests covering:

- new replay allowlist fields parse and default correctly
- outcome-specific lists are available to the replay gate

- [ ] **Step 2: Run the config tests and verify they fail**

Run the smallest relevant config test command.

Expected:

```text
FAIL
```

- [ ] **Step 3: Implement the config fields**

Add simple config-backed policy fields:

- `replay.allowed_actors`
- `replay.release_allowed_actors`
- `replay.refund_allowed_actors`

Ensure defaults are explicit and fail-closed.

- [ ] **Step 4: Re-run the config tests and verify they pass**

Run the same config test command again.

Expected:

```text
ok
```

- [ ] **Step 5: Commit the config slice**

Run:

```bash
git add <config files and tests>
git -c commit.gpgsign=false commit -m "feat: add replay allowlist config"
```

### Task 3: Publish Docs and Sync OpenSpec

**Files:**
- Create: `docs/architecture/policy-driven-replay-controls.md`
- Modify: `docs/architecture/operator-replay-manual-retry.md`
- Modify: `docs/architecture/index.md`
- Modify: `docs/architecture/p2p-knowledge-exchange-track.md`
- Modify: `zensical.toml`
- Modify: `openspec/specs/project-docs/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`
- Modify: `openspec/specs/meta-tools/spec.md`
- Create: `openspec/changes/archive/2026-04-23-policy-driven-replay-controls/**`

- [ ] **Step 1: Add the public architecture page**

Write `docs/architecture/policy-driven-replay-controls.md` describing:

- purpose and scope
- policy model
- actor resolution
- fail-closed behavior
- current limits

- [ ] **Step 2: Wire architecture landing, track, and nav**

Update:

- `docs/architecture/index.md`
- `docs/architecture/operator-replay-manual-retry.md`
- `docs/architecture/p2p-knowledge-exchange-track.md`
- `zensical.toml`

to reference the landed replay-policy slice truthfully.

- [ ] **Step 3: Sync OpenSpec main specs**

Update:

- `openspec/specs/project-docs/spec.md`
- `openspec/specs/docs-only/spec.md`
- `openspec/specs/meta-tools/spec.md`

to reflect:

- the new public page
- the landed track status
- the replay authorization contract

- [ ] **Step 4: Archive the completed change**

Create:

- `openspec/changes/archive/2026-04-23-policy-driven-replay-controls/proposal.md`
- `openspec/changes/archive/2026-04-23-policy-driven-replay-controls/design.md`
- `openspec/changes/archive/2026-04-23-policy-driven-replay-controls/tasks.md`
- delta spec stubs under `specs/`

and mark the change as complete.

- [ ] **Step 5: Run verification and commit docs/OpenSpec closeout**

Run:

```bash
.venv/bin/zensical build
go build ./...
go test ./...
```

Expected:

```text
all pass
```

Then commit:

```bash
git add docs/architecture/policy-driven-replay-controls.md docs/architecture/operator-replay-manual-retry.md docs/architecture/index.md docs/architecture/p2p-knowledge-exchange-track.md zensical.toml openspec/specs/project-docs/spec.md openspec/specs/docs-only/spec.md openspec/specs/meta-tools/spec.md openspec/changes/archive/2026-04-23-policy-driven-replay-controls
git -c commit.gpgsign=false commit -m "specs: archive policy driven replay controls"
```

---

## Sequencing Notes

- Task 1 first:
  - replay-service policy gate
- Task 2 second:
  - config-backed allowlist
- Task 3 third:
  - public docs
  - spec sync
  - archive closeout

## Verification Checklist

- [ ] `go test ./internal/postadjudicationreplay/... -count=1`
- [ ] focused config test command
- [ ] `.venv/bin/zensical build`
- [ ] `go build ./...`
- [ ] `go test ./...`

## Definition of Done

- replay is fail-closed when actor cannot be resolved
- replay is fail-closed when actor is not allowed for the current outcome
- config-backed allowlists exist for replay, release replay, and refund replay
- public docs and track page reflect the landed slice
- OpenSpec main specs are synced
- archived change exists under `openspec/changes/archive/2026-04-23-policy-driven-replay-controls`
- verification passes cleanly
