## MODIFIED Requirements

### Requirement: Doctor diagnostics coverage stays executable

Repository-level regressions in doctor diagnostics SHALL be enforced by executable tests.

#### Scenario: Doctor server port conflict message remains covered
- **WHEN** the doctor server port check receives a configured port that is already bound
- **THEN** executable tests SHALL fail if the diagnostic stops reporting that the port is in use
