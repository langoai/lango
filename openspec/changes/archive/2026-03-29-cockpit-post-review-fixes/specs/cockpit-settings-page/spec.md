## MODIFIED Requirements

### Requirement: Embedded settings with OnSave callback
SettingsPage SHALL embed `settings.Editor` via `NewEditorForEmbedding(cfg, onSave)`. The Editor SHALL work on a deep copy of the config, not the live runtime config. The save action SHALL pass context-related dotted path keys as explicitKeys, not category IDs.

#### Scenario: Save passes context-related explicit keys
- **WHEN** user selects "save" in embedded Editor
- **THEN** OnSave SHALL receive explicitKeys containing dotted paths from config.ContextRelatedKeys() (e.g., "knowledge.enabled"), not category keys (e.g., "knowledge")

#### Scenario: Form edits do not mutate live config
- **WHEN** user edits form fields in embedded Editor without saving
- **THEN** the original config passed to NewSettingsPage SHALL remain unchanged

#### Scenario: Auto-enable respects embedded save
- **WHEN** embedded save stores explicitKeys and config is reloaded
- **THEN** ResolveContextAutoEnable SHALL NOT override explicitly set values
