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
