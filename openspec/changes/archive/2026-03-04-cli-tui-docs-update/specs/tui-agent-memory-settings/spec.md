## ADDED Requirements

### Requirement: Agent memory settings form
The system SHALL provide a TUI settings form named "Agent Memory Configuration" accessible under the "AI & Knowledge" menu category. The form SHALL use tuicore.FormModel and include fields for agent memory configuration options.

#### Scenario: Form displays current values
- **WHEN** user navigates to Settings > AI & Knowledge > Agent Memory Configuration
- **THEN** the form displays current values for enabled, default scope, and confidence threshold

### Requirement: Agent memory form fields
The agent memory form SHALL include the following fields:
- `agent_memory_enabled` (InputBool): Enable/disable agent memory system
- `agent_memory_default_scope` (InputSelect): Default memory scope (instance/type/global)
- `agent_memory_min_confidence` (InputText): Minimum confidence threshold for memory retrieval (0.0-1.0)
- `agent_memory_max_entries` (InputInt): Maximum entries per agent before pruning

#### Scenario: Toggle enabled state
- **WHEN** user toggles the "Enabled" field via space key
- **THEN** the field value changes and is persisted when the form is saved

#### Scenario: Select default scope
- **WHEN** user cycles through scope options using left/right keys
- **THEN** the selected scope value updates among instance, type, and global

### Requirement: Agent memory form registration
The NewAgentMemoryForm() function SHALL be registered in the settings editor dispatch so it is accessible from the TUI settings menu.

#### Scenario: Settings menu includes agent memory
- **WHEN** user opens the TUI settings menu
- **THEN** "Agent Memory Configuration" appears under the "AI & Knowledge" category
