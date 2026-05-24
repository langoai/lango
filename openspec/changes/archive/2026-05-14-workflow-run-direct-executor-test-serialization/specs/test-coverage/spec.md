## ADDED Requirements

### Requirement: Workflow run schedule regressions remain deterministic
Workflow run command regressions that depend on package-global execution seams SHALL avoid parallel execution patterns that can leak stubbed state across tests.

#### Scenario: Direct-execution seam override does not affect sibling tests
- **WHEN** a workflow run regression temporarily replaces the package-global direct-execution seam
- **THEN** sibling workflow run regressions SHALL not observe that override unexpectedly
- **AND** repository-wide test runs SHALL remain deterministic
