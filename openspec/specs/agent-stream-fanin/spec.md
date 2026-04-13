## Purpose

AgentStreamFanIn provides a higher-level abstraction for merging multiple child agent output streams into a single tagged stream. It delegates to the stream-combinators Merge primitive and emits progress lifecycle events via ProgressBus, enabling supervisors to monitor child agent output in real time.

## Requirements

### Requirement: AgentStreamFanIn Type

The package SHALL provide an `AgentStreamFanIn` struct that holds a parent identifier, a map of child IDs to `Stream[string]`, and an optional `*ProgressBus`.

#### Scenario: Construction with bus
- **WHEN** `NewAgentStreamFanIn(parent, bus)` is called with a non-nil bus
- **THEN** it SHALL return an `AgentStreamFanIn` with the given parent ID and an empty children map

#### Scenario: Construction with nil bus
- **WHEN** `NewAgentStreamFanIn(parent, nil)` is called
- **THEN** it SHALL return an `AgentStreamFanIn` that operates without emitting progress events

### Requirement: AddChild Registration

The `AddChild` method SHALL register a child agent's output stream by child ID.

#### Scenario: Register a child stream
- **WHEN** `AddChild(childID, stream)` is called
- **THEN** the stream SHALL be stored in the children map keyed by childID

### Requirement: MergedStream Returns Tagged Stream

The `MergedStream` method SHALL return a `Stream[Tag[string]]` that yields all events from all registered child streams, tagged with their child ID as the source.

#### Scenario: Two children merged output
- **WHEN** two children are registered with streams producing ["a1","a2"] and ["b1","b2"]
- **THEN** `MergedStream` SHALL yield 4 tagged events with correct source attribution

#### Scenario: Single child degenerate case
- **WHEN** one child is registered
- **THEN** `MergedStream` SHALL yield all events from that child tagged with its ID

#### Scenario: Empty children returns empty stream
- **WHEN** no children are registered
- **THEN** `MergedStream` SHALL yield zero events

### Requirement: MergedStream Uses Merge Combinator

`MergedStream` SHALL delegate to the existing `Merge` combinator for stream merging. It SHALL NOT reimplement merge logic.

#### Scenario: Delegation to Merge
- **WHEN** `MergedStream` is called with N children
- **THEN** it SHALL construct a `map[string]Stream[string]` and pass it to `Merge`

### Requirement: Progress Lifecycle Emission

`MergedStream` SHALL emit progress events via `ProgressBus` for each child's lifecycle.

#### Scenario: ProgressStarted on merge begin
- **WHEN** `MergedStream` is called
- **THEN** it SHALL emit `ProgressStarted` for each registered child before yielding events

#### Scenario: ProgressCompleted on child finish
- **WHEN** a child stream completes without error
- **THEN** it SHALL emit `ProgressCompleted` for that child with Progress=1.0

#### Scenario: ProgressFailed on child error
- **WHEN** a child stream yields an error
- **THEN** it SHALL emit `ProgressFailed` for that child with the error in the message

#### Scenario: Progress source format
- **WHEN** a progress event is emitted for child "alpha" under parent "session-1"
- **THEN** the Source field SHALL be `"agent:session-1:child:alpha"`

### Requirement: Nil Bus Safety

All progress emission SHALL be skipped when the bus is nil. No nil pointer panics SHALL occur.

#### Scenario: Nil bus does not panic
- **WHEN** `AgentStreamFanIn` is created with nil bus and `MergedStream` is consumed
- **THEN** no panic SHALL occur and all events SHALL be yielded normally

### Requirement: Error Propagation

When a child stream yields an error, the error SHALL be propagated through the merged stream. Other children SHALL continue producing events.

#### Scenario: One child error others continue
- **WHEN** child-b yields an error after producing events
- **THEN** child-a's events SHALL still be present in the merged output
- **AND** the error SHALL be propagated through the merged stream
