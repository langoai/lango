## MODIFIED Requirements

### Requirement: Default entry point

Running `lango` (no subcommand) SHALL launch the mission workbench rather than the multi-panel cockpit TUI or the single-column chat TUI. The focused chat TUI SHALL remain accessible via `lango chat`, and the explicit cockpit shell SHALL remain accessible via `lango cockpit`.

#### Scenario: Default launches the mission workbench
- **WHEN** the user runs `lango` without a subcommand
- **THEN** the mission workbench SHALL launch
- **AND** bare `lango` SHALL no longer alias the cockpit shell

#### Scenario: Focused chat remains explicit
- **WHEN** the user runs `lango chat`
- **THEN** the single-column chat TUI SHALL launch with the same focused-chat behavior as before

#### Scenario: Explicit cockpit subcommand still launches cockpit
- **WHEN** the user runs `lango cockpit`
- **THEN** the multi-panel cockpit TUI SHALL launch
- **AND** it SHALL remain distinct from bare `lango`
