## ADDED Requirements

### Requirement: OpenTelemetry stdout trace exporter writer is seam-aware
When `observability.tracing.exporter` is `"stdout"`, the system SHALL construct
the stdout trace exporter with an explicit writer seam while preserving the
default stderr behavior.

#### Scenario: Stdout trace exporter uses injected writer
- **WHEN** the tracing writer seam is replaced for a test
- **AND** stdout tracing emits and flushes a span
- **THEN** the span SHALL be written through the injected writer
