## ADDED Requirements

### Requirement: Incremental bundle creation
The system SHALL create incremental git bundles using `base..HEAD` commit range when a valid base commit is provided. If the base commit does not exist in the repository, the system SHALL automatically fall back to creating a full bundle.

#### Scenario: Successful incremental bundle
- **WHEN** CreateIncrementalBundle is called with a valid base commit that exists in the repo
- **THEN** the system creates a bundle containing only commits after the base commit and returns the bundle bytes and HEAD hash

#### Scenario: Base commit not found triggers fallback
- **WHEN** CreateIncrementalBundle is called with a valid 40-char hex hash that does not exist in the repo
- **THEN** the system falls back to CreateBundle (full bundle) and returns the result

#### Scenario: Invalid base commit hash rejected
- **WHEN** CreateIncrementalBundle is called with a hash that is not 40 lowercase hex characters
- **THEN** the system returns an error indicating invalid base commit

### Requirement: Bundle verification before apply
The system SHALL verify bundle integrity and prerequisites before applying a bundle to a workspace repository.

#### Scenario: Valid bundle passes verification
- **WHEN** VerifyBundle is called with a valid bundle whose prerequisites exist in the repo
- **THEN** the system returns nil (no error)

#### Scenario: Bundle with missing prerequisites
- **WHEN** VerifyBundle is called with a bundle whose prerequisite commits are not in the repo
- **THEN** the system returns ErrMissingPrerequisite

### Requirement: Transactional bundle apply with rollback
The system SHALL snapshot all refs before applying a bundle and restore them if the apply fails.

#### Scenario: Successful safe apply
- **WHEN** SafeApplyBundle is called with a valid, verified bundle
- **THEN** the system verifies the bundle, snapshots refs, applies the bundle, and returns nil

#### Scenario: Rollback on apply failure
- **WHEN** SafeApplyBundle is called with a bundle that fails during unbundle
- **THEN** the system restores all refs to their pre-apply state and returns the apply error

### Requirement: Commit existence check
The system SHALL check whether a specific commit exists in a workspace repository.

#### Scenario: Commit exists
- **WHEN** HasCommit is called with a commit hash that exists in the repo
- **THEN** the system returns (true, nil)

#### Scenario: Commit does not exist
- **WHEN** HasCommit is called with a valid hash not present in the repo
- **THEN** the system returns (false, nil)

#### Scenario: Invalid hash
- **WHEN** HasCommit is called with an invalid commit hash format
- **THEN** the system returns an error
