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

## MODIFIED Requirements

### Requirement: Default entry point
Running `lango` (no subcommand) SHALL launch the multi-panel cockpit TUI instead of the single-column chat TUI. The single-column TUI SHALL remain accessible via `lango chat`.

#### Scenario: Default launches cockpit
- **WHEN** user runs `lango` without subcommand
- **THEN** the cockpit TUI SHALL launch (sidebar + main content + optional context panel)

#### Scenario: Legacy chat accessible
- **WHEN** user runs `lango chat`
- **THEN** the single-column chat TUI SHALL launch with the same behavior as the previous default

#### Scenario: Explicit cockpit subcommand
- **WHEN** user runs `lango cockpit`
- **THEN** the cockpit TUI SHALL launch, identical to bare `lango`
