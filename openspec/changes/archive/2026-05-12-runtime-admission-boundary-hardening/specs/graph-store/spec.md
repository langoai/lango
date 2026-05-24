## ADDED Requirements

### Requirement: Observe-only admission for supported runtime graph inputs
The Phase A1 graph store observe slice SHALL compute admission decisions only for the supported runtime graph inputs in this slice: event-bus `TriplesExtractedEvent` batches whose `Source` is `conversation_analysis`, `session_learning`, `learning`, or `proactive_librarian`, plus triples already returned from the `content.saved` extraction path. Observe-only telemetry for that extraction path SHALL use the synthetic `content_saved_extractor` source label. Observe mode SHALL NOT drop, rewrite, or reroute the original triple slice before graph write execution.

For this slice, `conversation_analysis`, `session_learning`, and `learning` are the supported learning-group producer sources, `proactive_librarian` is the supported librarian-group producer source, and `content_saved_extractor` is a separate synthetic telemetry source label for the extraction path.

For every observed triple slice in this requirement:
- `batch_count` SHALL equal `1`
- `known_count + unknown_count + unvalidated_count` SHALL equal the number of triples in that slice

#### Scenario: Event-bus triple producer source is observed
- **WHEN** a supported `TriplesExtractedEvent` producer-source batch is processed in observe mode
- **THEN** the runtime SHALL compute an observe-only admission decision for the original triple slice
- **AND** that decision SHALL classify the slice into `known_count`, `unknown_count`, and `unvalidated_count`
- **AND** the runtime SHALL preserve the original triples unchanged for graph write execution

#### Scenario: Content-saved extraction source is observed
- **WHEN** the `content.saved` extraction path returns triples in observe mode
- **THEN** the runtime SHALL compute an observe-only admission decision for the original extracted triple slice
- **AND** that decision SHALL classify the slice into `known_count`, `unknown_count`, and `unvalidated_count`
- **AND** the runtime SHALL preserve the original extracted triples unchanged for graph write execution

#### Scenario: Unsupported event-bus source keeps the existing write path
- **WHEN** observe mode receives a `TriplesExtractedEvent` batch whose `Source` is outside the supported event-bus producer-source set for this slice
- **THEN** the runtime SHALL skip observe-only admission classification for that batch
- **AND** the runtime SHALL preserve the original triples unchanged for graph write execution

#### Scenario: Validator unavailable keeps observe-only admission in unvalidated observation mode
- **WHEN** ontology is disabled or ontology initialization fails before observe-only admission can classify a supported runtime graph input
- **THEN** the runtime SHALL still compute one batch-scoped observe-only admission decision for that graph input
- **AND** all triples in that slice SHALL contribute to `unvalidated_count`
- **AND** `known_count` and `unknown_count` SHALL both equal `0`
- **AND** the decision SHALL use `validator_source = "unavailable"`
- **AND** the runtime SHALL preserve the original triple slice unchanged for graph write execution
