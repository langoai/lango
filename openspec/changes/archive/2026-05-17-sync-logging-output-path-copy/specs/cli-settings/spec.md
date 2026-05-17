## ADDED Requirements

### Requirement: Logging settings copy matches default stderr output

The Logging settings form SHALL describe an empty `logging.outputPath` as using the default stderr logging stream. It SHALL NOT describe an empty logging output path as stdout.

#### Scenario: Logging output path field describes stderr fallback

- **WHEN** the Logging settings form is rendered
- **THEN** the `log_output_path` placeholder and description SHALL communicate that an empty value uses stderr
- **AND** they SHALL NOT say an empty value uses stdout
