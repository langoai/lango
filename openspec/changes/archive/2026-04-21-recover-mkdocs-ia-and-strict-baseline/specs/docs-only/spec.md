## ADDED Requirements

### Requirement: Quickstart installation anchor resolves
The getting started quickstart documentation SHALL link to the existing installation anchor instead of a missing fragment.

#### Scenario: Installation anchor is valid
- **WHEN** a user reads `docs/getting-started/quickstart.md`
- **THEN** the installation link SHALL target the existing installation section and its compiler setup anchor

### Requirement: Cockpit public-entry consolidation
The public cockpit documentation SHALL consolidate the operator-facing material from the cockpit approval, channels, tasks, and troubleshooting sub-guides into `docs/features/cockpit.md`.

#### Scenario: Approval guidance is on the main cockpit page
- **WHEN** a user reads `docs/features/cockpit.md`
- **THEN** they SHALL find approval operations guidance previously split into the approval sub-guide

#### Scenario: Channel, task, and troubleshooting guidance are on the main cockpit page
- **WHEN** a user reads `docs/features/cockpit.md`
- **THEN** they SHALL find channel operations, background task operations, and troubleshooting guidance previously split into the other cockpit sub-guides
