## ADDED Requirements

### Requirement: Active approvals resolve through the shared pending response path
The TUI approval surfaces SHALL continue to use the existing approval tier rendering, but in cockpit mode the active approval request SHALL be owned by a shared pending approval registry that preserves the original response channel. Tier rendering must not create a second approval pipeline.

#### Scenario: Tier rendering uses the shared pending request in cockpit
- **WHEN** a pending approval is displayed from Mission Control or cockpit chat
- **THEN** both surfaces SHALL render from the same underlying pending approval request
- **AND** they SHALL present the same risk, reason, and action labels
- **AND** Slice 1 SHALL treat that shared request as the single live approval decision in cockpit state

#### Scenario: Approval resolves through the original response channel
- **WHEN** the user approves, denies, or allows for session from any cockpit TUI approval surface
- **THEN** the response SHALL be written to the original pending approval response channel
- **AND** no parallel pending-approval queue or history-only resolution path SHALL be introduced

### Requirement: Approval history remains completed-decision data
Completed approval history and grants MAY still inform detail surfaces, but they SHALL NOT be treated as the pending source of truth for the live decision in Mission Control.

#### Scenario: History does not satisfy a pending live decision
- **WHEN** Mission Control needs to render a currently pending approval
- **THEN** it SHALL read the shared pending approval owner instead of reconstructing the live decision from `approval.HistoryStore`

#### Scenario: History still usable for completed detail views
- **WHEN** a cockpit detail surface renders completed approvals or session grants
- **THEN** it MAY continue reading the history and grant stores as completed-decision data
