## MODIFIED Requirements

### Requirement: Workspace Isolation
Production runtime SHALL activate workspace isolation for coding steps once the execution-isolation stage is enabled. Error messages SHALL provide guided remediation with actionable commands. Patch application failures SHALL include rollback instructions.

#### Scenario: Runtime isolation active
- **WHEN** `runLedger.workspaceIsolation` is enabled
- **THEN** the app runtime wires `PEVEngine.WithWorkspace(...)`
- **AND** coding-step validators execute inside isolated worktrees rather than the base tree

#### Scenario: Retry-safe repeated validation
- **WHEN** the same step is validated multiple times under isolation
- **THEN** each attempt uses a retry-safe workspace identity
- **AND** previous attempts do not block later ones via reused branch metadata

#### Scenario: Dirty tree guided remediation
- **WHEN** `CheckDirtyTree` detects uncommitted changes
- **THEN** the error message includes a count of changed files
- **AND** the error suggests `git stash push -m "lango-workspace-isolation"` as a remediation command

#### Scenario: Patch apply conflict guidance
- **WHEN** `ApplyPatch` fails due to a merge conflict
- **THEN** the error message includes the raw git output
- **AND** the error instructs the user to run `git am --abort` to rollback

#### Scenario: Enablement conditions
- **WHEN** the system evaluates whether workspace isolation should be active
- **THEN** isolation is required for steps with validators of type `file_changed`, `build_pass`, or `test_pass`
- **AND** isolation is not required for validators of type `human_approval` or `always_pass`

## ADDED Requirements

### Requirement: RunLedger Workspace Isolation doctor check
The doctor command SHALL include a `RunLedger Workspace Isolation` check that validates the workspace isolation configuration and environment health. The check name SHALL be distinct from the existing `P2P Workspaces` check.

#### Scenario: Isolation enabled and healthy
- **WHEN** `runLedger.workspaceIsolation` is enabled and git is available and no stale worktrees exist
- **THEN** the check status is `Pass`
- **AND** the message includes the config value and active worktree count

#### Scenario: Isolation disabled
- **WHEN** `runLedger.workspaceIsolation` is disabled
- **THEN** the check status is `Skip`
- **AND** the message indicates isolation is not enabled

#### Scenario: Git unavailable
- **WHEN** `runLedger.workspaceIsolation` is enabled but `git` is not in PATH
- **THEN** the check status is `Warn`
- **AND** the message indicates git is required for workspace isolation

#### Scenario: Stale worktrees detected
- **WHEN** `git worktree list` reports worktrees under the runledger temp directory that no longer exist on disk
- **THEN** the check status is `Warn`
- **AND** the message lists the stale worktree paths

#### Scenario: Doctor help text
- **WHEN** user runs `lango doctor --help`
- **THEN** the output lists `RunLedger Workspace Isolation` under the Execution category
- **AND** the total check count is incremented by 1
