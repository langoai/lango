### Requirement: Observe-only admission for runtime dynamic producers
The Phase A1 runtime app observe slice SHALL compute admission decisions only for the current `TriplesExtractedEvent` producer set plus `content.saved` extractor outcomes already surfaced in app wiring, while preserving current write routing in observe mode.

#### Scenario: TriplesExtractedEvent observe path
- **WHEN** a `TriplesExtractedEvent` batch is processed
- **THEN** the runtime SHALL record graph-admission telemetry
- **AND** the runtime SHALL still enqueue the original triples

#### Scenario: Extracted triples observe path
- **WHEN** `content.saved` extraction returns triples
- **THEN** the runtime SHALL record observe-only admission telemetry for those returned triples
- **AND** the runtime SHALL still enqueue the original extracted triples

#### Scenario: Extractor dropped-unknown baseline
- **WHEN** the current `content.saved` extraction path rejects an unknown predicate before admission
- **THEN** the runtime SHALL record dropped-unknown baseline telemetry for that extractor path

#### Scenario: GraphBuffer write-failure baseline
- **WHEN** a batched graph write fails in observe mode
- **THEN** the runtime SHALL record a graph write-failure baseline event without changing existing error handling
