## MODIFIED Requirements

### Requirement: Default SaveableTools allowlist
The `toolchain` package SHALL define a `DefaultSaveableTools` constant as a fallback for CLI mode. At runtime, the `app.go` builder SHALL derive the saveable tools list from the tool catalog using `Catalog.SaveableToolNames()`, which filters by `ToolCapability.KnowledgeSaveable()`. The `BuildHookRegistry` function SHALL accept an optional catalog parameter and prefer catalog-derived tools when available.

#### Scenario: KnowledgeSaveHook uses catalog-derived list at runtime
- **WHEN** the app builder constructs `KnowledgeSaveHook` with a non-nil catalog
- **THEN** the hook's `SaveableTools` set equals the catalog's `SaveableToolNames()` result
- **AND** the list includes all tools where `ReadOnly == true` or `Activity ∈ {read, query}`

#### Scenario: KnowledgeSaveHook falls back to constant in CLI mode
- **WHEN** `BuildHookRegistry` is called with a nil catalog (CLI snapshot mode)
- **THEN** the hook's `SaveableTools` set equals `DefaultSaveableTools`

#### Scenario: CLI hooks output indicates source
- **WHEN** user runs `lango agent hooks`
- **THEN** the KnowledgeSaveHook details indicate whether the saveable tools list is "catalog-derived" or "fallback constant"
