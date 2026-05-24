## ADDED Requirements

### Requirement: Context panel renders unavailable messaging when metrics are absent
The cockpit context panel SHALL distinguish an unavailable metrics collector from valid zero-valued metric data.

#### Scenario: Nil metrics collector renders unavailable messages
- **WHEN** the context panel renders with no configured metrics collector
- **THEN** the Token Usage section SHALL explain that the metrics collector is not configured
- **AND** the Tool Stats section SHALL explain that the metrics collector is not configured
- **AND** the System section SHALL explain that the metrics collector is not configured
