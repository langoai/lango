## Context

`WorkspaceManager` in `runledger/workspace.go` handles git worktree isolation for coding steps. `CheckDirtyTree` runs `git status --porcelain` and returns a generic error. `ApplyPatch` runs `git am` and wraps the raw output. Both messages are developer-oriented, not user-friendly.

The existing `WorkspaceCheck` in `doctor/checks/workspace.go` is for **P2P Workspaces** — completely separate from RunLedger workspace isolation. A new check with a distinct name (`RunLedger Workspace Isolation`) avoids confusion.

`doctor.go` currently lists 26 checks; this adds one more → 27.

## Goals / Non-Goals

**Goals:**
- Guided remediation messages with actionable commands
- `git am` failure includes rollback instructions
- Doctor check for RunLedger workspace isolation health
- Spec delta documenting enablement conditions

**Non-Goals:**
- Cleanup telemetry persistence (P1 — no data source yet)
- Changing `lango run status` or `lango p2p workspace` CLI commands
- Enabling workspace isolation by default (P2 — after hardening)

## Decisions

### D1: `CheckDirtyTree` guided remediation

**Choice**: Run `git status --porcelain`, count changed files, include a suggested `git stash push -m "lango-workspace-isolation"` command in the error. Keep the error as a single `fmt.Errorf` — no new types.

### D2: `ApplyPatch` conflict guidance

**Choice**: On `git am` failure, wrap the error with explicit instructions: "To abort and return to the previous state: `git am --abort`". No automatic rollback — the user decides.

### D3: Doctor check uses `git worktree list`

**Choice**: Parse `git worktree list --porcelain` output to find active worktrees. Detect stale worktrees by checking if the worktree path still exists on disk. Git availability is checked via `exec.LookPath("git")`.

### D4: No cleanup telemetry

**Choice**: "Recent cleanup success/failure" is deferred to P1. The check covers: config value, git availability, active worktrees, stale detection. No new persistence needed.
