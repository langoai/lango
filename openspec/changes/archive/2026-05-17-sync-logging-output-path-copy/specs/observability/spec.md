## MODIFIED Requirements

### Requirement: Default logging output is stderr and seam-aware

When logging is initialized without an explicit writer or output path, the system SHALL write log entries through a package-level default writer seam that defaults to stderr. Explicit `LogConfig.Writer` and `LogConfig.OutputPath` SHALL keep precedence over the default seam. Public configuration documentation and settings copy SHALL describe an empty `logging.outputPath` as using stderr.

#### Scenario: Default logging does not use stdout

- **WHEN** `logging.Init(...)` is called without `Writer` or `OutputPath`
- **THEN** the default writer seam SHALL point at stderr
- **AND** stdout SHALL remain reserved for command result output

#### Scenario: Public configuration docs describe stderr fallback

- **WHEN** users read README or configuration documentation for `logging.outputPath`
- **THEN** they SHALL see that an empty value uses stderr by default
