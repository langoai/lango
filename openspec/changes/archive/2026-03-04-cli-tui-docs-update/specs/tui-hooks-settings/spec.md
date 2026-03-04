## ADDED Requirements

### Requirement: Hooks settings form
The system SHALL provide a TUI settings form named "Hooks Configuration" accessible under the "Communication" menu category. The form SHALL use tuicore.FormModel and include fields for all hooks configuration options.

#### Scenario: Form displays current values
- **WHEN** user navigates to Settings > Communication > Hooks Configuration
- **THEN** the form displays current values for enabled, securityFilter, accessControl, eventPublishing, knowledgeSave, and blockedCommands

### Requirement: Hooks form fields
The hooks form SHALL include the following fields:
- `hooks_enabled` (InputBool): Enable/disable the hook system
- `hooks_security_filter` (InputBool): Enable security filter hook
- `hooks_access_control` (InputBool): Enable per-agent tool access control hook
- `hooks_event_publishing` (InputBool): Enable event bus publishing hook
- `hooks_knowledge_save` (InputBool): Enable knowledge save hook
- `hooks_blocked_commands` (InputText): Comma-separated list of blocked command patterns

#### Scenario: Toggle hook enabled state
- **WHEN** user toggles the "Enabled" field via space key
- **THEN** the field value changes and is persisted when the form is saved

#### Scenario: Edit blocked commands
- **WHEN** user enters "rm -rf,shutdown" in the "Blocked Commands" field
- **THEN** the value is stored as a comma-separated string and parsed into a string slice on save

### Requirement: Hooks form registration
The NewHooksForm() function SHALL be registered in the settings editor dispatch so it is accessible from the TUI settings menu.

#### Scenario: Settings menu includes hooks
- **WHEN** user opens the TUI settings menu
- **THEN** "Hooks Configuration" appears under the "Communication" category
