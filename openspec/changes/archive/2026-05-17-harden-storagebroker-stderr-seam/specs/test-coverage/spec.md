## ADDED Requirements

### Requirement: Broker stderr seam regressions stay executable
Broker stderr routing regressions SHALL be enforced by executable tests instead
of relying only on manual review.

#### Scenario: Broker child stderr seam is covered
- **WHEN** the storage broker command is constructed
- **THEN** an executable test SHALL verify that child stderr uses the injected
  writer seam
