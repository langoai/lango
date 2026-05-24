## ADDED Requirements

### Requirement: Logging output path copy regressions stay executable

Repository-level regressions that reintroduce stale stdout fallback copy for `logging.outputPath` SHALL be enforced by executable tests.

#### Scenario: Logging settings form copy is covered

- **WHEN** settings form tests run
- **THEN** they SHALL fail if `log_output_path` copy says an empty value uses stdout
