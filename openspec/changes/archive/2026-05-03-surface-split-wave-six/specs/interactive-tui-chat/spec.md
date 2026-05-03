## MODIFIED Requirements

### Requirement: Interactive TUI chat on bare invocation

Running `lango chat` SHALL start the interactive terminal chat session using Bubble Tea. `lango serve` SHALL continue to work as the full gateway plus channels mode. Wave 6 SHALL remove the older bare-`lango` chat interpretation from this surface contract.

#### Scenario: Explicit chat command launches TUI chat
- **WHEN** the user runs `lango chat` on an interactive TTY
- **THEN** an interactive TUI chat session starts

#### Scenario: Bare `lango` no longer means direct chat
- **WHEN** the user runs bare `lango`
- **THEN** this specification SHALL NOT claim that the focused chat surface starts
- **AND** the bare-`lango` contract SHALL belong to the mission workbench instead

#### Scenario: `lango serve` is unchanged
- **WHEN** the user runs `lango serve`
- **THEN** the full gateway plus channels server starts as before
