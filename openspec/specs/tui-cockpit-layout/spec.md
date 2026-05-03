## Purpose

Define the single-column coding-agent cockpit layout for the interactive `lango` TUI.
## Requirements
### Requirement: Single-column cockpit regions
The interactive TUI SHALL render as a single-column coding-agent cockpit with four primary regions: header, turn status strip, transcript viewport, and footer.

#### Scenario: Default idle layout
- **WHEN** the user runs `lango` and the TUI enters idle state
- **THEN** the screen SHALL show a header, a turn status strip, a transcript viewport, and a footer in that top-to-bottom order

#### Scenario: Streaming layout
- **WHEN** the agent is actively streaming a response
- **THEN** the same primary regions SHALL remain visible and the turn status strip SHALL indicate that generation is in progress

### Requirement: Approval interrupt card
Approval requests SHALL be rendered as interrupt cards within the single-column layout instead of opening separate modal or side panels.

#### Scenario: Approval request shown inline
- **WHEN** a tool approval request is raised during a turn
- **THEN** the TUI SHALL display an approval interrupt card in the main column with the tool name, summary, key parameters, and action keys

#### Scenario: Approval result retained in transcript
- **WHEN** an approval request is approved or denied
- **THEN** the transcript SHALL retain a compact approval event entry describing the outcome

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

