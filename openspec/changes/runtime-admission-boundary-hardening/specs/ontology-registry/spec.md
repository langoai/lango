### Requirement: Shared predicate validity source
The runtime SHALL use the ontology service predicate validator closure as the shared predicate-validity source for observe-only admission decisions and graph-store validation.

#### Scenario: Validator source is surfaced
- **WHEN** observe-only admission telemetry is recorded
- **THEN** the runtime SHALL tag the telemetry with the validator source identifier used for predicate checks
