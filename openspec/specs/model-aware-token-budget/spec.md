## ADDED Requirements

### Requirement: Model-family-aware token budgeting
The system SHALL provide a `ModelTokenBudget(modelName)` function that returns an appropriate history token budget based on the model family's context window size.

#### Scenario: Claude models
- **WHEN** the model name contains "claude" (case-insensitive)
- **THEN** the budget SHALL be 100,000 tokens (~50% of 200K context)

#### Scenario: Gemini models
- **WHEN** the model name contains "gemini" (case-insensitive)
- **THEN** the budget SHALL be 200,000 tokens (~20% of 1M context)

#### Scenario: GPT-4o and GPT-4-turbo models
- **WHEN** the model name contains "gpt-4o" or "gpt-4-turbo" (case-insensitive)
- **THEN** the budget SHALL be 64,000 tokens (~50% of 128K context)

#### Scenario: GPT-4 base models
- **WHEN** the model name contains "gpt-4" but not "gpt-4o" or "gpt-4-turbo"
- **THEN** the budget SHALL be 32,000 tokens

#### Scenario: GPT-3.5 models
- **WHEN** the model name contains "gpt-3.5" (case-insensitive)
- **THEN** the budget SHALL be 8,000 tokens (~50% of 16K context)

#### Scenario: Unknown model fallback
- **WHEN** the model name does not match any known family
- **THEN** the budget SHALL be the DefaultTokenBudget (32,000 tokens)

### Requirement: Token budget propagation through session service
The `SessionServiceAdapter` SHALL propagate a configured token budget to all `SessionAdapter` instances it creates, which in turn pass it to `EventsAdapter` for history truncation.

#### Scenario: WithTokenBudget sets budget on adapter
- **WHEN** `WithTokenBudget(budget)` is called on the session service
- **THEN** all subsequently created sessions SHALL use that budget for history truncation

### Requirement: Lazy caching of truncated history and events
The `EventsAdapter` SHALL lazily compute and cache truncated history and converted events using `sync.Once` for O(1) repeated access.

#### Scenario: Multiple calls to truncatedHistory
- **WHEN** `truncatedHistory()` is called multiple times
- **THEN** the token-budget truncation SHALL execute only once; subsequent calls return the cached result

#### Scenario: Multiple calls to At
- **WHEN** `At(i)` is called for different indices
- **THEN** the full event list SHALL be built once on first `At()` call and cached for subsequent calls

#### Scenario: Out-of-bounds At access
- **WHEN** `At(i)` is called with `i < 0` or `i >= len(events)`
- **THEN** the method SHALL return nil
