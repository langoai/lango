## ADDED Requirements

### Requirement: TTY Guard for TUI Commands
The root command, `cockpit` subcommand, and `chat` subcommand SHALL detect whether stdin is an interactive terminal before launching the TUI. Non-interactive environments MUST NOT attempt to start bubbletea.

#### Scenario: Root command in non-TTY environment
- **WHEN** `lango` is invoked without an interactive terminal (e.g., piped stdin, CI, `</dev/null`)
- **THEN** the command SHALL print help text and exit with code 0

#### Scenario: Cockpit subcommand in non-TTY environment
- **WHEN** `lango cockpit` is invoked without an interactive terminal
- **THEN** the command SHALL return an error: "cockpit requires an interactive terminal"

#### Scenario: Chat subcommand in non-TTY environment
- **WHEN** `lango chat` is invoked without an interactive terminal
- **THEN** the command SHALL return an error: "chat requires an interactive terminal"

#### Scenario: Normal interactive execution
- **WHEN** `lango` is invoked in an interactive terminal
- **THEN** the cockpit TUI SHALL launch normally (no behavior change)
