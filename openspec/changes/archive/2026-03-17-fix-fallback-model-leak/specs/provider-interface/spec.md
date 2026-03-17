## MODIFIED Requirements

### Requirement: Gemini provider runtime model validation
The Gemini provider's `Generate()` method SHALL call `ValidateModelProvider("gemini", model)` after alias normalization and before making the API call.

#### Scenario: Wrong model routed to Gemini at runtime
- **WHEN** `Generate()` receives `params.Model = "gpt-5.3-codex"`
- **THEN** it SHALL return an error wrapping `ErrModelProviderMismatch` without making an API call

#### Scenario: Valid Gemini model passes runtime check
- **WHEN** `Generate()` receives `params.Model = "gemini-3-flash-preview"`
- **THEN** the validation SHALL pass and the API call SHALL proceed

### Requirement: Anthropic provider runtime model validation
The Anthropic provider's `Generate()` method SHALL call `ValidateModelProvider("anthropic", params.Model)` before processing the request.

#### Scenario: Wrong model routed to Anthropic at runtime
- **WHEN** `Generate()` receives `params.Model = "gpt-5.3-codex"`
- **THEN** it SHALL return an error wrapping `ErrModelProviderMismatch` without making an API call

#### Scenario: Valid Anthropic model passes runtime check
- **WHEN** `Generate()` receives `params.Model = "claude-sonnet-4-5-20250514"`
- **THEN** the validation SHALL pass and the request SHALL proceed
