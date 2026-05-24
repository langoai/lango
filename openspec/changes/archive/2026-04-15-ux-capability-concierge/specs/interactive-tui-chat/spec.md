## MODIFIED Requirements

### Requirement: Slash commands
The TUI SHALL support slash commands: `/help`, `/clear`, `/new`, `/model`, `/status`, `/exit`, `/quit`, `/mode`, `/cost`. The `/mode` command SHALL accept a mode name argument and update the session's mode accordingly; without an argument, it SHALL print the current mode and available modes. The `/cost` command SHALL print the session's cumulative token usage and estimated cost.

#### Scenario: /clear resets chat
- **WHEN** the user types `/clear`
- **THEN** the chat viewport is cleared and a new session starts

#### Scenario: /mode with name sets session mode
- **WHEN** the user types `/mode code-review`
- **THEN** the session mode SHALL be set to `code-review`
- **AND** a `ModeChangedEvent` SHALL be published
- **AND** the chat view SHALL display a system status entry indicating the new mode

#### Scenario: /mode without argument lists modes
- **WHEN** the user types `/mode`
- **THEN** the TUI SHALL print the current mode (or "none") and the list of available modes

#### Scenario: /cost prints session summary
- **WHEN** the user types `/cost` after turns have occurred
- **THEN** the TUI SHALL print cumulative input tokens, output tokens, and estimated cost
