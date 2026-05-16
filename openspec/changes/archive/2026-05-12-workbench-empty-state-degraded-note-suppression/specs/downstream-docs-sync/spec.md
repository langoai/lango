## ADDED Requirements

### Requirement: Public workbench docs describe calmer empty-state warning behavior

Public workbench documentation SHALL explain that the standalone empty workbench avoids showing cockpit-style degraded warnings before any active mission/control context exists.

#### Scenario: Docs mention empty-state warning suppression
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL mention that the empty workbench emphasizes setup and next actions instead of surfacing cockpit degraded warnings immediately
