## MODIFIED Requirements

### Requirement: TTY Guard for TUI Commands

The root command, `cockpit` subcommand, and `chat` subcommand SHALL detect whether stdin is an interactive terminal before launching their respective TUI surfaces. Non-interactive environments MUST NOT attempt to start Bubble Tea.

#### Scenario: Root command in non-TTY environment
- **WHEN** `lango` is invoked without an interactive terminal
- **THEN** the command SHALL print help text and exit with code 0

#### Scenario: Cockpit subcommand in non-TTY environment
- **WHEN** `lango cockpit` is invoked without an interactive terminal
- **THEN** the command SHALL return an error: "cockpit requires an interactive terminal"

#### Scenario: Chat subcommand in non-TTY environment
- **WHEN** `lango chat` is invoked without an interactive terminal
- **THEN** the command SHALL return an error: "chat requires an interactive terminal"

#### Scenario: Explicit cockpit launch in interactive execution
- **WHEN** `lango cockpit` is invoked in an interactive terminal
- **THEN** the cockpit TUI SHALL launch normally

### Requirement: Cockpit root model orchestrates Mission Control as the default first cockpit surface

The cockpit root model SHALL treat Mission Control as the default active page for explicit `lango cockpit` launches while continuing to host the existing detail pages. The sidebar remains secondary navigation inside cockpit, but cockpit no longer owns the bare-`lango` contract in Wave 6.

#### Scenario: Explicit cockpit launch starts on Mission Control
- **WHEN** cockpit is created for the explicit `lango cockpit` surface
- **THEN** `activePage` SHALL initialize to `PageMissionControl`

#### Scenario: Chat remains a detail page inside explicit cockpit
- **WHEN** the user navigates from Mission Control to Chat inside cockpit
- **THEN** the chat child surface SHALL remain reachable as a detail route
- **AND** `lango chat` SHALL still bypass cockpit Mission Control entirely
