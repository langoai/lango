## Why

The current P2P git bundle system creates full-repository bundles on every sync, which becomes prohibitively expensive as workspaces grow. Additionally, multiple agents modifying the same branch simultaneously cause merge conflicts with no automated resolution. These two issues—bandwidth waste and conflict chaos—are the primary bottlenecks for scaling P2P agent collaboration.

## What Changes

- **Incremental bundles**: `CreateIncrementalBundle(base..HEAD)` sends only new commits since a known base, with automatic fallback to full bundles when the base commit is missing
- **Transactional bundle apply**: `SafeApplyBundle` verifies bundle prerequisites, snapshots refs before applying, and rolls back on failure
- **Bundle verification**: `VerifyBundle` validates bundle integrity and prerequisites before applying
- **Branch-per-task isolation**: Each agent task gets its own `task/{taskID}` branch to eliminate conflicts at the source
- **Bare-repo merge**: `MergeTaskBranch` uses `git merge-tree --write-tree` + `commit-tree` + `update-ref` for 3-way merges on bare repos without a working tree
- **Conflict detection**: Structured `MergeResult` with conflict file list when merge fails
- **Health monitor git state tracking**: Health pings now collect HEAD commit hashes, enabling divergence detection across team members
- **New workspace message types**: `CONFLICT_REPORT`, `BRANCH_CREATED`, `BRANCH_MERGED`, `SYNC_REQUEST` for branch collaboration signaling
- **Config extensions**: `EnableIncrementalBundle`, `BranchPerTask`, `GitStateTracking`, `AutoSyncOnDivergence` flags

## Capabilities

### New Capabilities
- `incremental-git-bundle`: Incremental bundle creation, verification, transactional apply with rollback, and ref snapshot/restore
- `branch-per-task`: Task branch lifecycle (create, list, merge, delete) with bare-repo merge-tree support and conflict detection
- `git-state-tracking`: Health monitor extension for tracking member HEAD hashes and detecting workspace git divergence

### Modified Capabilities
- `p2p-gitbundle`: New protocol message types (push_incremental_bundle, fetch_incremental, verify_bundle, has_commit) and handler dispatch
- `p2p-workspace`: New message types for conflict and branch collaboration signaling
- `p2p-team-coordination`: Health monitor extended with git state provider and divergence detection

## Impact

- **Code**: `internal/p2p/gitbundle/` (bundle.go, messages.go, protocol.go, branch.go new), `internal/p2p/team/health_monitor.go`, `internal/p2p/workspace/message.go`, `internal/config/types_p2p.go`, `internal/eventbus/workspace_events.go`
- **Protocol**: 4 new request types added to git bundle protocol (backward compatible — old clients ignore unknown types)
- **Config**: 5 new config fields with safe defaults (all disabled/zero-value by default)
- **Dependencies**: No new external dependencies; uses existing git CLI and go-git
