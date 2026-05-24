## MODIFIED Requirements

### Requirement: Cockpit public-entry consolidation
After the hidden cockpit guides move out of `docs/`, the public cockpit documentation SHALL keep `docs/features/cockpit.md` as the single public entry for operator-facing material from the cockpit approval, channels, tasks, and troubleshooting guides.

#### Scenario: Approval guidance is on the main cockpit page
- **WHEN** a user reads `docs/features/cockpit.md`
- **THEN** they SHALL find approval operations guidance previously split into the approval sub-guide

#### Scenario: Channel, task, and troubleshooting guidance are on the main cockpit page
- **WHEN** a user reads `docs/features/cockpit.md`
- **THEN** they SHALL find channel operations, background task operations, and troubleshooting guidance previously split into the other cockpit sub-guides
