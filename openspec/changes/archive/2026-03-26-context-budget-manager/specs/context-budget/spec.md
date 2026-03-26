## ADDED Requirements

### Requirement: Model window registry
The system SHALL provide a `LookupModelWindow(modelName string) int` function that returns the context window size in tokens for known models. The lookup SHALL match model name prefixes (e.g., "gemini-2.0" matches "gemini-2.0-flash-001"). If no match is found, the system SHALL return a configurable default (128k tokens).

#### Scenario: Known model lookup
- **WHEN** `LookupModelWindow("gemini-2.0-flash")` is called
- **THEN** the system SHALL return 1000000 (1M tokens)

#### Scenario: Known model prefix match
- **WHEN** `LookupModelWindow("gpt-4o-2024-08-06")` is called
- **THEN** the system SHALL match the "gpt-4o" prefix and return 128000

#### Scenario: Unknown model fallback
- **WHEN** `LookupModelWindow("custom-model-v1")` is called
- **THEN** the system SHALL return the default window size (128000)

#### Scenario: Config override
- **WHEN** `context.modelWindow` is set to a positive value in config
- **THEN** the configured value SHALL be used instead of the registry lookup

### Requirement: ContextBudgetManager construction and validation
The system SHALL provide a `ContextBudgetManager` type that accepts model window, response reserve, base prompt tokens, and a `SectionAllocation`. The allocation sum (knowledge + rag + memory + runSummary + headroom) MUST equal 1.0. Construction SHALL fail with an error if the sum is not 1.0 (within floating-point tolerance of 0.001).

#### Scenario: Valid allocation
- **WHEN** a `ContextBudgetManager` is created with allocation summing to 1.0
- **THEN** construction SHALL succeed without error

#### Scenario: Invalid allocation sum
- **WHEN** a `ContextBudgetManager` is created with allocation summing to 0.8
- **THEN** construction SHALL return an error containing "allocation sum"

#### Scenario: Response reserve clamping
- **WHEN** response reserve is 0
- **THEN** the system SHALL use `Agent.MaxTokens` as the default (4096)
- **AND** the reserve SHALL be clamped to at least 1024 and at most 25% of model window

### Requirement: Budget calculation
The `ContextBudgetManager` SHALL compute per-section token budgets using the formula: `available = modelWindow - responseReserve - basePromptTokens`, then `sectionBudget = available * allocation[section]`. The `basePromptTokens` SHALL be estimated from the base system prompt using `types.EstimateTokens()`.

#### Scenario: Standard budget computation
- **WHEN** model window is 128000, response reserve is 4096, base prompt is 2000 tokens
- **THEN** available is 121904
- **AND** knowledge budget is 121904 * 0.30 = 36571 tokens
- **AND** RAG budget is 121904 * 0.25 = 30476 tokens
- **AND** memory budget is 121904 * 0.25 = 30476 tokens
- **AND** runSummary budget is 121904 * 0.10 = 12190 tokens

#### Scenario: Small model budget
- **WHEN** model window is 8192, response reserve is 4096, base prompt is 3000 tokens
- **THEN** available is 1096
- **AND** section budgets SHALL be proportionally small but positive

### Requirement: Graceful degradation on zero or negative available budget
When `available <= 0` (model window too small or base prompt too large), the `ContextBudgetManager` SHALL return unlimited (0) budgets for all sections, effectively disabling budget enforcement. The system SHALL log a warning.

#### Scenario: Negative available budget
- **WHEN** model window is 4096, response reserve is 4096, base prompt is 2000 tokens
- **THEN** available is -2000
- **AND** all section budgets SHALL be 0 (unlimited)
- **AND** a warning SHALL be logged

#### Scenario: Zero available budget
- **WHEN** model window equals response reserve plus base prompt tokens exactly
- **THEN** all section budgets SHALL be 0 (unlimited)

### Requirement: Budget-aware section assembly
The `ContextAwareModelAdapter.GenerateContent()` SHALL compute per-section budgets from the `ContextBudgetManager` (if set) before parallel retrieval. Each section assembly SHALL receive its budget and truncate content to fit. When no budget manager is set, existing unbounded behavior SHALL be preserved.

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

### Requirement: Config surface for context budget
The system SHALL provide a `ContextConfig` struct in `internal/config/types.go` with fields: `ModelWindow` (int), `ResponseReserve` (int), and `Allocation` (SectionAllocation with Knowledge, RAG, Memory, RunSummary, Headroom float64 fields). Default allocation SHALL be 0.30/0.25/0.25/0.10/0.10. `contextProfile` SHALL remain top-level and independent.

#### Scenario: Default config
- **WHEN** no `context.*` config is set
- **THEN** default allocation (0.30/0.25/0.25/0.10/0.10) SHALL be used
- **AND** model window SHALL be auto-detected from model registry

#### Scenario: Custom allocation
- **WHEN** `context.allocation.knowledge` is set to 0.40 and other values adjusted
- **THEN** the custom allocation SHALL be used if the sum equals 1.0

#### Scenario: contextProfile independence
- **WHEN** `contextProfile: balanced` is set alongside `context.allocation.*`
- **THEN** both SHALL work independently (profile controls feature enables, context controls budgets)
