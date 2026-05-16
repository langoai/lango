## ADDED Requirements

### Requirement: Sidebar reflects cockpit page availability
The cockpit sidebar SHALL visually disable optional pages that are not currently registered in the running cockpit model instead of presenting them as live navigation targets.

#### Scenario: Unregistered optional pages start disabled
- **WHEN** a cockpit model is created before optional pages such as Tools, Status, Sessions, Tasks, Dead Letters, or Approvals are registered
- **THEN** the corresponding sidebar items SHALL be marked disabled
- **AND** core routes Mission Control and Chat SHALL remain enabled

#### Scenario: Registering a page enables its sidebar item
- **WHEN** a cockpit page such as Tools or Tasks is registered on the model
- **THEN** the matching sidebar item SHALL become enabled
