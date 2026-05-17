## MODIFIED Requirements

### Requirement: Settings command requires an interactive terminal
The `lango settings` command SHALL fail before bootstrap or TUI startup when the current stdin is not an interactive terminal.

#### Scenario: Non-interactive settings fails with scripted guidance
- **WHEN** `lango settings` is invoked while stdin is not an interactive terminal
- **THEN** the command SHALL return an error that says settings requires an interactive terminal
- **AND** the error SHALL guide scripted configuration toward `lango config import` or `lango config set`
- **AND** the command SHALL NOT start the settings editor or save a profile
