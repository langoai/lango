## ADDED Requirements

### Requirement: Provider Interface Definition
The system SHALL define a `Provider` interface that all LLM provider implementations MUST implement.

#### Scenario: Interface methods defined
- **WHEN** a new LLM provider is implemented
- **THEN** it SHALL implement `ID() string` returning the provider identifier
- **AND** SHALL implement `Generate(ctx, params) iter.Seq2[StreamEvent, error]` for streaming responses
- **AND** SHALL implement `ListModels(ctx) ([]ModelInfo, error)` for model discovery

### Requirement: Streaming Response Support
The system SHALL support streaming LLM responses via Go iterators.

#### Scenario: Text streaming
- **WHEN** the provider generates a response
- **THEN** it SHALL yield `StreamEvent` with `Type: "text_delta"` for each text chunk

#### Scenario: Tool call streaming
- **WHEN** the provider generates a tool call
- **THEN** it SHALL yield `StreamEvent` with `Type: "tool_call"` containing the tool call details

#### Scenario: Thought text streaming
- **WHEN** the provider generates thought-only text (e.g., Gemini `Thought=true`)
- **THEN** it SHALL yield `StreamEvent` with `Type: "thought"` and `ThoughtLen` set to the byte length of the thought text
- **AND** it SHALL NOT include the thought text content in the `Text` field

#### Scenario: Stream completion
- **WHEN** the response generation completes
- **THEN** it SHALL yield `StreamEvent` with `Type: "done"`

#### Scenario: Error during streaming
- **WHEN** an error occurs during generation
- **THEN** it SHALL yield `StreamEvent` with `Type: "error"` and the error details

### Requirement: Generation Parameters
The system SHALL accept standard generation parameters across all providers.

#### Scenario: Common parameters supported
- **WHEN** `GenerateParams` is passed to a provider
- **THEN** it SHALL accept `Model`, `Messages`, `Tools`, `Temperature`, and `MaxTokens`

### Requirement: Model Information
The system SHALL provide standardized model metadata.

#### Scenario: ModelInfo structure
- **WHEN** `ListModels` is called
- **THEN** each `ModelInfo` SHALL contain `ID`, `Name`, `ContextWindow`, `SupportsVision`, `SupportsTools`, and `IsReasoning` fields

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

### Requirement: ToolCall carries provider-specific metadata
The `provider.ToolCall` struct SHALL include `Thought bool` and `ThoughtSignature []byte` fields to support Gemini thinking metadata passthrough. These fields SHALL be zero-valued for non-Gemini providers.

#### Scenario: Gemini FunctionCall with ThoughtSignature
- **WHEN** a Gemini streaming response contains a FunctionCall part with `Thought=true` and `ThoughtSignature` set
- **THEN** the resulting `provider.ToolCall` SHALL have `Thought=true` and `ThoughtSignature` populated with the original bytes

#### Scenario: Non-Gemini provider ToolCall
- **WHEN** a non-Gemini provider emits a ToolCall
- **THEN** the `Thought` field SHALL be `false` and `ThoughtSignature` SHALL be `nil`

### Requirement: Thought event type for provider streaming
The `StreamEventType` enum SHALL include a `StreamEventThought` value (`"thought"`) for thought-only text filtered at the provider level. The `StreamEvent` struct SHALL include a `ThoughtLen int` field carrying the byte length of filtered thought text for diagnostics.

#### Scenario: StreamEventThought is a valid event type
- **WHEN** `StreamEventThought.Valid()` is called
- **THEN** it SHALL return `true`

#### Scenario: StreamEventThought included in Values()
- **WHEN** `StreamEventType.Values()` is called
- **THEN** the returned slice SHALL include `StreamEventThought`

#### Scenario: ThoughtLen populated on thought events
- **WHEN** a provider emits a `StreamEventThought` event
- **THEN** the `ThoughtLen` field SHALL contain the byte length of the filtered thought text
- **AND** the `Text` field SHALL be empty (thought content is not exposed)
