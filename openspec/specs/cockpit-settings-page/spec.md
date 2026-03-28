## ADDED Requirements

### Requirement: Embedded settings with OnSave callback
SettingsPage SHALL embed `settings.Editor` via `NewEditorForEmbedding(cfg, onSave)`. The `"save"` menu action SHALL call OnSave instead of tea.Quit when OnSave is non-nil.

#### Scenario: Save in embedded mode
- **WHEN** user selects "save" in embedded Editor with OnSave set
- **THEN** Editor SHALL call OnSave(cfg, explicitKeys) and return to menu without tea.Quit

#### Scenario: Save in standalone mode
- **WHEN** user selects "save" in standalone Editor (OnSave=nil)
- **THEN** Editor SHALL set Completed=true and return tea.Quit (existing behavior)

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
