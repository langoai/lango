## ADDED Requirements

### Requirement: Default logging output is stderr and seam-aware

When logging is initialized without an explicit writer or output path, the system SHALL write log entries through a package-level default writer seam that defaults to stderr. Explicit `LogConfig.Writer` and `LogConfig.OutputPath` SHALL keep precedence over the default seam.

#### Scenario: Default logging uses injected writer seam

- **WHEN** `logging.Init(...)` is called without `Writer` or `OutputPath`
- **AND** a test replaces the default logging writer seam
- **THEN** emitted log entries SHALL be written through the injected seam
- **AND** no process-global stdout interception SHALL be required

#### Scenario: Default logging does not use stdout

- **WHEN** `logging.Init(...)` is called without `Writer` or `OutputPath`
- **THEN** the default writer seam SHALL point at stderr
- **AND** stdout SHALL remain reserved for command result output
