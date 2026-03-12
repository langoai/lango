## ADDED Requirements

### Requirement: Branch collaboration message types
The workspace message system SHALL support four new message types for branch-based collaboration signaling.

#### Scenario: Conflict report message
- **WHEN** a merge conflict occurs between branches
- **THEN** a CONFLICT_REPORT message is posted with metadata containing conflictFiles, sourceBranch, targetBranch, sourceAgent, taskID, and resolution fields

#### Scenario: Branch created message
- **WHEN** a task branch is created in a workspace
- **THEN** a BRANCH_CREATED message is posted to notify other workspace members

#### Scenario: Branch merged message
- **WHEN** a task branch is successfully merged
- **THEN** a BRANCH_MERGED message is posted to notify other workspace members

#### Scenario: Sync request message
- **WHEN** git state divergence is detected or a member requests synchronization
- **THEN** a SYNC_REQUEST message is posted to coordinate re-sync
