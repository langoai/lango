## MODIFIED Requirements

### Requirement: Dead Letters remains registered as a cockpit degraded surface
The cockpit SHALL keep Dead Letters available as a registered page route even when the dead-letter bridge is unavailable, relying on page-level unavailable messaging instead of suppressing page registration.

#### Scenario: Retry-running state hides inert reset help
- **WHEN** the Dead Letters help is rendered while a retry request is actively running
- **THEN** it SHALL omit the `Ctrl+R` reset binding

#### Scenario: Retry-running state hides inert reset copy in the filter hint line
- **WHEN** the Dead Letters filter hint line is rendered while a retry request is actively running
- **THEN** it SHALL omit `Ctrl+R` reset wording
