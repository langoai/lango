## MODIFIED Requirements

### Requirement: Onboard command requires an interactive terminal
The `lango onboard` command SHALL fail before bootstrap or TUI startup when the current stdin is not an interactive terminal.

#### Scenario: Non-interactive onboard fails with scripted guidance
- **WHEN** `lango onboard` is invoked while stdin is not an interactive terminal
- **THEN** the command SHALL return an error that says onboard requires an interactive terminal
- **AND** the error SHALL guide scripted setup toward `lango config create --preset <name>` or `lango config import`
- **AND** the command SHALL NOT start the onboard wizard or save a profile
