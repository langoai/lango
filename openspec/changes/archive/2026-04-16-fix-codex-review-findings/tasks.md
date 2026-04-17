## 1. Fix 1 (P1): KnowledgeSaver nil on runtime path

- [x] 1.1 Add `knowledgeSaver toolchain.KnowledgeSaver` param to `buildHookRegistry` in `internal/app/app.go`
- [x] 1.2 At call site `app.go:177`, thread `knowledgeSaver` variable (nil for now — knowledge.Store does not yet implement KnowledgeSaver) with comment explaining the wiring path
- [x] 1.3 Existing tests pass (buildHookRegistry signature updated)

## 2. Fix 2 (P2): Baseline includes current run

- [x] 2.1 In `bootstrap_timing.go`, after `ReadTimingLog()`, exclude last entry from baseline with single-writer assumption comment
- [x] 2.2 Existing tests pass

## 3. Fix 3 (P2): git worktree list failure → false Pass

- [x] 3.1 Change `listRunLedgerWorktrees` return to `(active, stale []string, err error)`
- [x] 3.2 Propagate `cmd.Output()` error and `scanner.Err()` as returned error
- [x] 3.3 In `Run()`, map non-nil err to `StatusWarn`
- [x] 3.4 Existing tests pass

## 4. Fix 4 (P2): drift counter time boundary

- [x] 4.1 Replace `driftCounters map[string]int` with `map[string]driftEntry` (count + firstSeen)
- [x] 4.2 In `EmitSpecDrift`, reset counter if `now - firstSeen >= dedupWindow`
- [x] 4.3 In `Prune()`, iterate `driftCounters` directly by `firstSeen` instead of using `recentHashes` keys
- [x] 4.4 Existing drift tests pass

## 5. Fix 5 (P2): stash hint missing untracked

- [x] 5.1 Change `git stash push -m` to `git stash push -u -m` in `workspace.go`
- [x] 5.2 Existing test assertion still valid (substring match)

## 6. Fix 6: spec signature

- [x] 6.1 Delta spec created with 4-arg `BuildHookRegistry(cfg, bus, knowledgeSaver, catalog)` + runtime saver scenario

## 7. Verification

- [x] 7.1 `go build ./...` passes
- [x] 7.2 `go test ./...` passes — zero FAIL
