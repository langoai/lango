## MODIFIED Requirements

### Requirement: Materialized Snapshots
The system SHALL return deep copies of cached snapshots to prevent concurrent mutation. `GetRunSnapshot` MUST NOT return a pointer to the internal cache entry.

#### Scenario: Concurrent snapshot reads are isolated
- **WHEN** two goroutines call `GetRunSnapshot` for the same run concurrently
- **THEN** each receives an independent deep copy
- **AND** mutations to one copy do not affect the other or the cache

#### Scenario: Deep copy covers all mutable fields
- **WHEN** `RunSnapshot.DeepCopy()` is called
- **THEN** the returned snapshot has independent copies of `Steps`, `AcceptanceState`, `Notes`, and all nested slices/maps (`Evidence`, `DependsOn`, `ToolProfile`, `Validator.Params`, `MetAt`)
- **AND** modifying any field on the copy does not affect the original

#### Scenario: ApplyTail operates on copy in MemoryStore
- **WHEN** `MemoryStore.GetRunSnapshot` finds a cached snapshot with pending tail events
- **THEN** it calls `DeepCopy()` before passing the snapshot to `ApplyTail`
- **AND** the cache is updated with the mutated copy after tail application succeeds

#### Scenario: ApplyTail operates on copy in EntStore
- **WHEN** `EntStore.GetRunSnapshot` finds a cached snapshot with pending tail events
- **THEN** it calls `DeepCopy()` before passing the snapshot to `ApplyTail`
- **AND** the cache is updated with the mutated copy after tail application succeeds

### Requirement: PEV Engine
`VerifyAcceptanceCriteria` SHALL NOT mutate its input slice. It SHALL return both the list of unmet criteria and a fully evaluated copy of the criteria slice.

#### Scenario: Input criteria slice is not mutated
- **GIVEN** a criteria slice where all items have `Met = false`
- **WHEN** `VerifyAcceptanceCriteria` is called and some criteria pass validation
- **THEN** the original criteria slice items still have `Met = false`
- **AND** the returned evaluated copy has `Met = true` for passing criteria

#### Scenario: MetAt is set on newly met criteria
- **WHEN** a criterion passes validation in `VerifyAcceptanceCriteria`
- **THEN** the evaluated copy has `Met = true` and `MetAt` set to the current time via `time.Now()`
- **AND** `MetAt` is not nil

#### Scenario: Dead ctxKeyNow code is removed
- **WHEN** `VerifyAcceptanceCriteria` sets `MetAt` on a passing criterion
- **THEN** it uses `time.Now()` directly
- **AND** no context-value lookup for time injection exists

### Requirement: Run Lifecycle
`checkRunCompletion` SHALL only journal `EventCriterionMet` for criteria that are **newly** met — criteria that transitioned from `Met=false` to `Met=true` during the current verification pass.

#### Scenario: Already-met criteria are not re-journaled
- **GIVEN** a run where criterion 0 was already met (journaled in a previous pass)
- **WHEN** `checkRunCompletion` runs and criterion 1 is newly met
- **THEN** only one `EventCriterionMet` journal entry is appended (for criterion 1)
- **AND** no duplicate entry is created for criterion 0

#### Scenario: First-time met criteria are journaled
- **GIVEN** a run where no criteria have been met yet
- **WHEN** `checkRunCompletion` runs and criteria 0 and 2 pass validation
- **THEN** exactly two `EventCriterionMet` journal entries are appended (for indices 0 and 2)

## ADDED Requirements

### Requirement: Snapshot Deep Copy
The system SHALL provide a `RunSnapshot.DeepCopy()` method that returns a fully independent copy of the snapshot with no shared mutable state.

#### Scenario: DeepCopy produces independent snapshot
- **WHEN** `DeepCopy()` is called on a snapshot with steps, acceptance criteria, and notes
- **THEN** the returned snapshot has the same field values
- **AND** appending to `copy.Steps` does not affect the original's `Steps` slice
- **AND** modifying `copy.Notes["key"]` does not affect the original's `Notes` map

#### Scenario: DeepCopy preserves MetAt pointer semantics
- **GIVEN** a snapshot with an `AcceptanceCriterion` where `MetAt` points to a time value
- **WHEN** `DeepCopy()` is called
- **THEN** the copy's `MetAt` is a new pointer with the same time value
- **AND** modifying the copy's `MetAt` does not affect the original

### Requirement: Marshal Payload Error Observability
`marshalPayload` SHALL log a warning when JSON marshaling fails instead of silently returning an empty object.

#### Scenario: Marshal failure is logged
- **WHEN** `marshalPayload` receives a value that fails `json.Marshal` (e.g., channel type, cyclic reference)
- **THEN** a warning-level log message is emitted containing the error details
- **AND** the function still returns `{}` as a fallback
- **AND** the caller's flow is not interrupted

#### Scenario: Successful marshal is not logged
- **WHEN** `marshalPayload` receives a valid serializable value
- **THEN** no log message is emitted
- **AND** the correct JSON bytes are returned

### Requirement: Projection Sync Error Observability
Projection sync errors in `writethrough.go` SHALL be logged at warning level instead of being discarded with `_ =`.

#### Scenario: Degraded projection sync error is logged
- **WHEN** `appendProjectionSyncEvent` returns an error in a write-through method
- **THEN** a warning-level log message is emitted containing the run ID and error details
- **AND** the outer operation continues (best-effort semantics preserved)

#### Scenario: Successful projection sync is not logged
- **WHEN** `appendProjectionSyncEvent` returns nil
- **THEN** no warning log message is emitted
