## MODIFIED Requirements

### Requirement: Checkpoint Creation
The system SHALL support creating provenance checkpoints as thin metadata records referencing RunLedger journal positions. Checkpoints SHALL contain: ID, session_key, run_id, label, trigger type, journal_seq, optional git_ref, optional metadata, and created_at timestamp.

The internal `create()` method SHALL accept an optional `metadata map[string]string` parameter. When metadata is non-nil, it SHALL be set on the checkpoint before saving.

A new public method `CreateManualWithMetadata` SHALL be provided that accepts metadata and does NOT require a runID (runID is optional). The method SHALL return `ErrInvalidLabel` if label is empty.

#### Scenario: Manual checkpoint creation
- **WHEN** a user creates a checkpoint with a label and run ID
- **THEN** the system creates a checkpoint with trigger "manual" and the current journal seq for that run

#### Scenario: Manual checkpoint with metadata and no runID
- **WHEN** `CreateManualWithMetadata` is called with a label and metadata but empty runID
- **THEN** the system creates a checkpoint with trigger "manual", empty runID, and the provided metadata

#### Scenario: Manual checkpoint with metadata and empty label
- **WHEN** `CreateManualWithMetadata` is called with an empty label
- **THEN** the system returns `ErrInvalidLabel`

#### Scenario: Automatic checkpoint on step validation
- **WHEN** a RunLedger step validation passes and `provenance.checkpoints.autoOnStepComplete` is true
- **THEN** the system automatically creates a checkpoint with trigger "step_complete"

#### Scenario: Automatic checkpoint on policy applied
- **WHEN** a RunLedger policy decision is applied and `provenance.checkpoints.autoOnPolicy` is true
- **THEN** the system automatically creates a checkpoint with trigger "policy_applied"

#### Scenario: Max checkpoints per session enforcement
- **WHEN** a session has reached `provenance.checkpoints.maxPerSession` checkpoints
- **THEN** the system SHALL reject new checkpoint creation with ErrMaxCheckpoints

#### Scenario: Empty label rejected
- **WHEN** a checkpoint creation is attempted with an empty label
- **THEN** the system SHALL return ErrInvalidLabel
