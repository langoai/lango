## Context

Six issues from Codex review, all in code introduced during the agent-orchestration-roadmap branch.

## Decisions

Each fix is minimal and scoped to the specific defect. No architectural changes.

### Fix 1: Thread KnowledgeSaver through buildHookRegistry
Add `knowledgeSaver` param to private `buildHookRegistry`. At the call site in `app.go:177`, extract saver from `intelligenceValues.KC.store` if available. Nil-safe: if IV or KC is nil, pass nil (CLI/snapshot path remains unchanged).

### Fix 2: Exclude current run from timing baseline
After `ReadTimingLog()` in the check, drop the last entry (the current run appended by `Pipeline.Execute` before doctor runs). Single-writer assumption — document with a comment.

### Fix 3: Propagate git worktree list errors
Change `listRunLedgerWorktrees` to return `(active, stale, err)`. Propagate both `cmd.Output()` errors and `scanner.Err()`. Caller maps non-nil err to `StatusWarn`.

### Fix 4: Time-bound drift counters
Replace `map[string]int` with `map[string]driftEntry{count, firstSeen}`. On each call, if `now - firstSeen >= dedupWindow`, reset the counter (stale accumulation). `Prune()` iterates `driftCounters` directly by `firstSeen`, not through `recentHashes` keys.

### Fix 5: Include untracked in stash hint
Change `git stash push -m` to `git stash push -u -m`. One character.

### Fix 6: Spec signature alignment
Delta spec updating `BuildHookRegistry` requirement to 4-arg `(cfg, bus, knowledgeSaver, catalog)`.
