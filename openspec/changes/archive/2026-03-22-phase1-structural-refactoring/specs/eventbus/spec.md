## MODIFIED Requirements

### Requirement: ContentSavedEvent includes graph routing field
`ContentSavedEvent` SHALL include a `NeedsGraph bool` field that controls whether graph triple extraction runs for this event. Publishers MUST set `NeedsGraph` according to the original callback semantics: `true` for new knowledge creation and memory observations/reflections, `false` for knowledge updates and learning saves.

#### Scenario: New knowledge creation sets NeedsGraph true
- **WHEN** knowledge store creates a new entry
- **THEN** `ContentSavedEvent` is published with `IsNew: true, NeedsGraph: true`

#### Scenario: Knowledge update sets NeedsGraph false
- **WHEN** knowledge store updates an existing entry
- **THEN** `ContentSavedEvent` is published with `IsNew: false, NeedsGraph: false`

#### Scenario: Learning save sets NeedsGraph false
- **WHEN** knowledge store saves a learning entry
- **THEN** `ContentSavedEvent` is published with `IsNew: true, NeedsGraph: false`

#### Scenario: Memory observation sets NeedsGraph true
- **WHEN** memory store saves an observation
- **THEN** `ContentSavedEvent` is published with `IsNew: true, NeedsGraph: true`

### Requirement: ReputationChangedEvent for reputation store
A new `ReputationChangedEvent` SHALL be published by the reputation store when a peer's score changes, replacing `SetOnChangeCallback`.

#### Scenario: Reputation update publishes event
- **WHEN** reputation store updates a peer score
- **THEN** `ReputationChangedEvent{PeerDID, NewScore}` is published on the EventBus
