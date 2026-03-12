## Context

The P2P git bundle system (`internal/p2p/gitbundle/`) manages bare git repositories per workspace. Currently, `CreateBundle` always bundles all refs (`--all`), and `ApplyBundle` directly unbundles without verification or rollback. When multiple agents push to the same branch, conflicts arise with no resolution path.

Bare repos lack a working tree, so standard `git merge` is unavailable. Git 2.38+ introduced `merge-tree --write-tree` which performs 3-way merge producing a tree object, enabling merges without checkout.

## Goals / Non-Goals

**Goals:**
- Reduce bundle sizes by 90%+ for incremental syncs between agents with shared history
- Enable safe bundle application with verification and atomic rollback on failure
- Isolate agent work via per-task branches to eliminate concurrent modification conflicts
- Detect git state divergence across team members via health monitoring
- Provide structured conflict reports when merges fail

**Non-Goals:**
- Automatic conflict resolution (agents cannot meaningfully resolve code conflicts)
- File-level locking via economy system (excessive complexity)
- Custom merge strategies beyond Git's default recursive
- Working-tree operations (all operations target bare repos)

## Decisions

### 1. Incremental Bundles via `base..HEAD` Range
**Decision**: Use git's native range syntax for incremental bundles with automatic fallback to full bundles.
**Rationale**: Git bundle format natively supports commit ranges. The fallback ensures reliability when the base commit is missing (e.g., after repo compaction or first sync).
**Alternative considered**: Delta compression at application level — rejected because git's built-in range support is simpler and battle-tested.

### 2. Ref Snapshot/Restore for Transactional Apply
**Decision**: Capture all refs via `git for-each-ref` before applying, restore via `git update-ref` on failure.
**Rationale**: Bare repos don't support `git stash`. Ref snapshots provide equivalent rollback capability with minimal overhead.
**Alternative considered**: Copy entire .git directory — rejected due to I/O cost for large repos.

### 3. `git merge-tree --write-tree` for Bare-Repo Merge
**Decision**: Use merge-tree + commit-tree + update-ref pipeline for merging branches in bare repos.
**Rationale**: This is the only way to perform 3-way merges without a working tree. Available in Git 2.38+ which is widely deployed.
**Alternative considered**: Temporary working tree checkout — rejected because it requires disk space and cleanup, and is not safe for concurrent operations.

### 4. Branch-Per-Task Isolation
**Decision**: Each agent task gets a `task/{taskID}` branch, merged into target when complete.
**Rationale**: Eliminates conflicts at the source. Agents work independently; conflicts only surface at merge time when they can be reported structurally.
**Alternative considered**: File locking via economy tokens — rejected as overly complex with poor ergonomics.

### 5. Health Monitor Git State via Ping Extension
**Decision**: Extend health ping responses to include HEAD hashes, with majority-vote divergence detection.
**Rationale**: Lightweight addition to existing health check infrastructure. Divergence detection identifies sync issues without additional protocol messages.

## Risks / Trade-offs

- **[Git 2.38+ dependency]** → `merge-tree --write-tree` requires Git 2.38+. Mitigated by graceful error handling; branch operations still work, only merge is affected.
- **[Incremental bundle prerequisite miss]** → Remote may not have the base commit. Mitigated by `VerifyBundle` check and automatic fallback to full bundle.
- **[Ref snapshot size]** → Repos with many refs could have large snapshots. Mitigated by ref count being typically small (tens, not thousands) in workspace repos.
- **[Merge-tree false conflicts]** → Git may report conflicts that could be auto-resolved with more context. Mitigated by structured `MergeResult` that includes conflict file list for human/agent review.
