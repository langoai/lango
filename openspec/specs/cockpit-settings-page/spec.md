## Purpose

Capability spec for cockpit-settings-page. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Inline save result banner
Editor SHALL display save success or failure inline at menu top. The banner SHALL clear on next key input.

#### Scenario: Save success banner
- **WHEN** OnSave returns nil
- **THEN** Editor SHALL display a green "Settings saved" banner at menu top

#### Scenario: Save failure banner
- **WHEN** OnSave returns an error
- **THEN** Editor SHALL display a red error banner at menu top

### Requirement: Skip welcome step in embedded mode
`NewEditorForEmbedding()` SHALL create an Editor starting at StepMenu, skipping StepWelcome.

#### Scenario: Embedded editor starts at menu
- **WHEN** NewEditorForEmbedding(cfg, onSave) is called
- **THEN** the Editor step SHALL be StepMenu, not StepWelcome


### Requirement: Embedded settings with OnSave callback
SettingsPage SHALL embed `settings.Editor` via `NewEditorForEmbedding(cfg, onSave)`. The Editor SHALL work on a deep copy of the config, not the live runtime config. The save action SHALL pass context-related dotted path keys as explicitKeys, not category IDs.

#### Scenario: Save in embedded mode
- **WHEN** user selects "save" in embedded Editor with OnSave set
- **THEN** Editor SHALL call OnSave(cfg, explicitKeys) and return to menu without tea.Quit

#### Scenario: Save in standalone mode
- **WHEN** user selects "save" in standalone Editor (OnSave=nil)
- **THEN** Editor SHALL set Completed=true and return tea.Quit (existing behavior)

#### Scenario: Save passes context-related explicit keys
- **WHEN** user selects "save" in embedded Editor
- **THEN** OnSave SHALL receive explicitKeys containing dotted paths from config.ContextRelatedKeys(), not category keys

#### Scenario: Form edits do not mutate live config
- **WHEN** user edits form fields in embedded Editor without saving
- **THEN** the original config passed to NewSettingsPage SHALL remain unchanged

#### Scenario: Auto-enable respects embedded save
- **WHEN** embedded save stores explicitKeys and config is reloaded
- **THEN** ResolveContextAutoEnable SHALL NOT override explicitly set values
