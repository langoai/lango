## ADDED Requirements

### Requirement: Focused chat docs describe setup readiness gating
Public CLI documentation SHALL explain that focused chat uses the same setup readiness contract as the workbench.

#### Scenario: Core docs explain chat setup-required guidance
- **WHEN** a user reads `docs/cli/core.md`
- **THEN** the focused chat documentation SHALL state that incomplete profiles show setup guidance before normal turns
- **AND** the documentation SHALL mention `lango onboard`, `lango settings`, and `lango doctor`
