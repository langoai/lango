## ADDED Requirements

### Requirement: Lazy step index for FindStep
`RunSnapshot.FindStep(stepID)` SHALL use a lazily-built `map[string]int` index for O(1) lookup. The index SHALL be invalidated (set to nil) when `Steps` is mutated by `EventPlanAttached` or `PolicyDecompose`. The index SHALL NOT be serialized to JSON (`json:"-"` tag). After JSON unmarshal or `DeepCopy()`, the index SHALL be nil and rebuilt on next `FindStep` call.

#### Scenario: FindStep after PlanAttached
- **WHEN** a snapshot is materialized from journal events including `plan_attached`
- **THEN** `FindStep` SHALL return the correct step by ID in O(1) time

#### Scenario: FindStep after PolicyDecompose adds new steps
- **WHEN** `PolicyDecompose` appends new steps to the snapshot
- **THEN** `FindStep` SHALL find both original and newly-added steps

#### Scenario: FindStep after DeepCopy
- **WHEN** `DeepCopy()` is called and the original's steps are mutated
- **THEN** the copy's `FindStep` SHALL return its own independent step data

#### Scenario: FindStep after JSON round-trip
- **WHEN** a snapshot is marshaled to JSON and unmarshaled back
- **THEN** `FindStep` SHALL work correctly via lazy index rebuild

### Requirement: SourceKind and SourceDescriptor in RunCreatedPayload
`RunCreatedPayload` SHALL include `SourceKind string` (values: "workflow", "background", "") and `SourceDescriptor json.RawMessage` (original workflow or origin JSON). These fields SHALL be persisted in the journal and restored into `RunSnapshot` during event replay.

#### Scenario: Workflow run stores source descriptor
- **WHEN** `WorkflowWriteThrough.CreateRun()` creates a run
- **THEN** the journal event SHALL include `SourceKind: "workflow"` and the workflow marshaled as `SourceDescriptor`

#### Scenario: Background task stores source descriptor
- **WHEN** `BackgroundWriteThrough.PrepareTask()` creates a run
- **THEN** the journal event SHALL include `SourceKind: "background"` and the origin marshaled as `SourceDescriptor`

#### Scenario: Legacy journals without SourceKind
- **WHEN** a journal event from before this change is replayed
- **THEN** `SourceKind` SHALL be empty string and `SourceDescriptor` SHALL be nil (zero values)

## MODIFIED Requirements

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

#### Scenario: DeepCopy produces independent SourceDescriptor
- **WHEN** `DeepCopy()` is called on a snapshot with a non-nil `SourceDescriptor`
- **THEN** modifying the copy's `SourceDescriptor` backing array SHALL NOT affect the original

#### Scenario: DeepCopy resets step index
- **WHEN** `DeepCopy()` is called on a snapshot with a warm step index
- **THEN** the copy's step index SHALL be nil (lazy rebuild on next FindStep)
