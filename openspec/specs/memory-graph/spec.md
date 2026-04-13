## Purpose

Capability spec for memory-graph. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Memory GraphHooks for temporal triples
The memory system SHALL generate graph triples for temporal ordering and session membership when observations and reflections are saved.

#### Scenario: Observation creates session membership triple
- **WHEN** an observation is saved
- **THEN** an `in_session` triple SHALL be created: `observation:{ID} --[in_session]--> session:{SessionKey}`

#### Scenario: Observation creates temporal ordering triple
- **WHEN** an observation is saved and a previous observation exists for the same session
- **THEN** a `follows` triple SHALL be created: `observation:{ID} --[follows]--> observation:{PreviousID}`

#### Scenario: Reflection creates observation link triples
- **WHEN** a reflection is saved
- **THEN** `in_session` and `reflects_on` triples SHALL be created linking the reflection to its session and to all observations in that session

### Requirement: GraphHooks wiring in memory store
The memory Store SHALL accept a `SetGraphHooks(*GraphHooks)` method that enables triple generation on save operations.

#### Scenario: GraphHooks not set
- **WHEN** no GraphHooks are configured
- **THEN** SaveObservation and SaveReflection SHALL work normally without generating triples

#### Scenario: GraphHooks set
- **WHEN** GraphHooks are set via SetGraphHooks
- **THEN** SaveObservation SHALL call OnObservation and SaveReflection SHALL call OnReflection

### Requirement: Previous observation tracking
The memory Store SHALL track the last observation ID per session key to enable temporal ordering triples.

#### Scenario: First observation in session
- **WHEN** the first observation for a session is saved
- **THEN** no `follows` triple SHALL be created (no previous observation)

#### Scenario: Subsequent observations
- **WHEN** a second observation for the same session is saved
- **THEN** a `follows` triple SHALL link it to the first observation

### Requirement: Nil callback safety
GraphHooks SHALL safely handle nil callbacks without panicking.

#### Scenario: Nil callback on OnObservation
- **WHEN** OnObservation is called with a nil callback
- **THEN** the method SHALL return without error or panic
