## 1. Config & Message Types

- [x] 1.1 Add EnableIncrementalBundle, BranchPerTask, MaxIncrementalBundleSizeBytes to WorkspaceConfig
- [x] 1.2 Add GitStateTracking, AutoSyncOnDivergence to TeamConfig
- [x] 1.3 Add CONFLICT_REPORT, BRANCH_CREATED, BRANCH_MERGED, SYNC_REQUEST message types to workspace/message.go

## 2. Incremental Bundle & Transactional Apply

- [x] 2.1 Add ErrEmptyRepo, ErrMissingPrerequisite sentinel errors
- [x] 2.2 Implement validateCommitHash helper (40-char hex validation)
- [x] 2.3 Implement HasCommit (go-git CommitObject lookup)
- [x] 2.4 Implement CreateIncrementalBundle (base..HEAD with auto-fallback to full)
- [x] 2.5 Implement VerifyBundle (git bundle verify via temp file)
- [x] 2.6 Implement snapshotRefs / restoreRefs for ref-level rollback
- [x] 2.7 Implement SafeApplyBundle (verify + snapshot + apply + rollback on failure)

## 3. Protocol Extensions

- [x] 3.1 Add push_incremental_bundle, fetch_incremental, verify_bundle, has_commit request types and payloads
- [x] 3.2 Add handler dispatch cases in StreamHandler switch
- [x] 3.3 Implement handlePushIncrementalBundle, handleFetchIncremental, handleVerifyBundle, handleHasCommit

## 4. Branch-Per-Task Management

- [x] 4.1 Create branch.go with BranchInfo, MergeResult types and TaskBranchPrefix constant
- [x] 4.2 Implement CreateTaskBranch (idempotent, supports base branch)
- [x] 4.3 Implement MergeTaskBranch (merge-tree --write-tree + commit-tree + update-ref)
- [x] 4.4 Implement ListBranches (for-each-ref with parsed output)
- [x] 4.5 Implement DeleteTaskBranch (idempotent)
- [x] 4.6 Implement parseConflictFiles and resolveRef helpers

## 5. Health Monitor Git State Tracking

- [x] 5.1 Add GitStateProvider, MemberGitState, GitDivergence types
- [x] 5.2 Extend HealthMonitor struct with gitState, gitStateProv, workspaceIDs fields
- [x] 5.3 Extend HealthMonitorConfig with GitStateProvider, WorkspaceIDsFn
- [x] 5.4 Implement updateGitState and DetectDivergence methods
- [x] 5.5 Modify pingMember to collect git state after successful ping
- [x] 5.6 Add WorkspaceGitDivergenceEvent to eventbus/workspace_events.go

## 6. Tests

- [x] 6.1 Add incremental bundle tests (validateCommitHash, HasCommit, CreateIncrementalBundle, VerifyBundle, snapshotRefs, SafeApplyBundle)
- [x] 6.2 Add branch management tests (CreateTaskBranch, DeleteTaskBranch, ListBranches, parseConflictFiles, seedCommit helper)
- [x] 6.3 Add health monitor git state tests (GitStateTracking, DetectDivergence_NoMembers, DetectDivergence_AllSame, UpdateGitState_EmptyHash)
- [x] 6.4 Run full project test suite (go test ./... -count=1)
