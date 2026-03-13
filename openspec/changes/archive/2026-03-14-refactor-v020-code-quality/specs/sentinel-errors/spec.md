## MODIFIED Requirements

### Requirement: Alert metadata structure
The sentinel `Alert` type SHALL use a typed `AlertMetadata` struct instead of `map[string]interface{}` for the `Metadata` field. The struct SHALL include fields: `Count`, `Window`, `Amount`, `Threshold`, `Elapsed`, `PreviousBalance`, `NewBalance` with `json:"...,omitempty"` tags.

#### Scenario: Detector populates typed metadata
- **WHEN** `RapidCreationDetector` generates an alert for excessive deal creation
- **THEN** the alert's `Metadata.Count` and `Metadata.Window` fields are populated as typed values

#### Scenario: JSON marshaling preserves omitempty
- **WHEN** an alert has only `Count` and `Window` set in metadata
- **THEN** JSON output omits `amount`, `threshold`, `elapsed`, `previousBalance`, `newBalance` fields

### Requirement: Shared window counter for detectors
The sentinel package SHALL provide a `windowCounter` struct that encapsulates sliding-window counting logic shared by `RapidCreationDetector` and `RepeatedDisputeDetector`.

#### Scenario: windowCounter records and counts
- **WHEN** `record(key)` is called multiple times within the configured window
- **THEN** it returns the count of entries within the window, pruning expired entries

#### Scenario: Detectors embed windowCounter
- **WHEN** `RapidCreationDetector` and `RepeatedDisputeDetector` are initialized
- **THEN** both embed `windowCounter` instead of duplicating sliding-window logic

## ADDED Requirements

### Requirement: Domain error sentinels for escrow
The `internal/economy/escrow/` package SHALL export `ErrNotFunded` and `ErrInvalidStatus` sentinel errors.

#### Scenario: ErrNotFunded matching
- **WHEN** an escrow operation fails because funds are not deposited
- **THEN** the returned error wraps `ErrNotFunded` and is matchable via `errors.Is`

### Requirement: Domain error sentinel for workspace
The `internal/p2p/workspace/` package SHALL export `ErrWorkspaceNotFound`.

#### Scenario: Workspace lookup failure
- **WHEN** `manager.GetWorkspace(id)` is called with a non-existent workspace ID
- **THEN** the returned error wraps `ErrWorkspaceNotFound`

### Requirement: Domain error sentinel for workflow
The `internal/cli/workflow/` package SHALL export `ErrWorkflowDisabled`.

#### Scenario: Workflow operations when disabled
- **WHEN** any workflow function is called while the workflow engine is not enabled
- **THEN** it returns `ErrWorkflowDisabled`
