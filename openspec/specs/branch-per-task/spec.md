## Purpose

Capability spec for branch-per-task. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Task branch creation
The system SHALL create isolated `task/{taskID}` branches in bare workspace repositories. The operation SHALL be idempotent — creating an already-existing branch returns success.

#### Scenario: Create task branch from default HEAD
- **WHEN** CreateTaskBranch is called with a taskID and empty baseBranch
- **THEN** the system creates refs/heads/task/{taskID} pointing to the current HEAD

#### Scenario: Create task branch from specific base
- **WHEN** CreateTaskBranch is called with a taskID and baseBranch "main"
- **THEN** the system creates refs/heads/task/{taskID} pointing to the tip of "main"

#### Scenario: Idempotent creation
- **WHEN** CreateTaskBranch is called for a taskID whose branch already exists
- **THEN** the system returns nil without modifying the existing branch

#### Scenario: Empty task ID rejected
- **WHEN** CreateTaskBranch is called with an empty taskID
- **THEN** the system returns an error

### Requirement: Task branch merge via merge-tree
The system SHALL merge task branches into a target branch using `git merge-tree --write-tree` for bare-repo compatibility. On conflict, the system SHALL return a MergeResult with Success=false and the list of conflicting files.

#### Scenario: Clean merge
- **WHEN** MergeTaskBranch is called and no conflicts exist
- **THEN** the system creates a merge commit and updates the target ref, returning MergeResult with Success=true and the merge commit hash

#### Scenario: Merge with conflicts
- **WHEN** MergeTaskBranch is called and conflicting changes exist
- **THEN** the system returns MergeResult with Success=false, ConflictFiles populated, and no ref updates

### Requirement: Branch listing
The system SHALL list all branches in a workspace repository with their name, commit hash, HEAD status, and updated timestamp.

#### Scenario: List branches in repo with branches
- **WHEN** ListBranches is called on a repo with branches
- **THEN** the system returns a BranchInfo slice with name, commitHash, isHead, and updatedAt for each branch

#### Scenario: List branches in empty repo
- **WHEN** ListBranches is called on an empty repo
- **THEN** the system returns an empty slice and nil error

### Requirement: Task branch deletion
The system SHALL delete task branches. The operation SHALL be idempotent — deleting a non-existent branch returns success.

#### Scenario: Delete existing branch
- **WHEN** DeleteTaskBranch is called for an existing task branch
- **THEN** the system removes refs/heads/task/{taskID}

#### Scenario: Delete non-existent branch
- **WHEN** DeleteTaskBranch is called for a branch that does not exist
- **THEN** the system returns nil (idempotent)
