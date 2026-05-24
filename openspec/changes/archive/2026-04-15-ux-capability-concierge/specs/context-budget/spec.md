## MODIFIED Requirements

### Requirement: Budget-aware section assembly
The `ContextAwareModelAdapter.GenerateContent()` SHALL compute per-section budgets from the `ContextBudgetManager` (if set) before parallel retrieval. Each section assembly SHALL receive its budget and truncate content to fit. When no budget manager is set, existing unbounded behavior SHALL be preserved. The tool catalog section SHALL be generated dynamically per turn from the injected `*toolcatalog.Catalog` (not from `basePrompt`), using mode-filtered listing when a session mode is active.

#### Scenario: Budget manager set
- **WHEN** `WithBudgetManager()` has been called with a valid budget manager
- **THEN** `GenerateContent()` SHALL compute budgets and pass them to section assembly methods

#### Scenario: Budget manager not set
- **WHEN** no budget manager is configured
- **THEN** `GenerateContent()` SHALL assemble sections without budget limits (legacy behavior)

#### Scenario: Section within budget
- **WHEN** a section's content fits within its allocated budget
- **THEN** the section SHALL be included in full, with no truncation

#### Scenario: Section exceeds budget
- **WHEN** a section's content exceeds its allocated budget
- **THEN** the section SHALL be truncated by dropping lower-priority items until within budget

#### Scenario: Tool catalog generated dynamically
- **WHEN** `GenerateContent()` executes with a `WithCatalog()` wired catalog
- **THEN** the tool catalog section SHALL be regenerated from the catalog for this turn
- **AND** the section SHALL reflect the session's active mode allowlist if a mode is set
