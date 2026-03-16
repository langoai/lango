## ADDED Requirements

### Requirement: OpenAI API Compatibility
The system SHALL implement a provider that uses the OpenAI Chat Completions API format.

#### Scenario: Standard OpenAI endpoint
- **WHEN** provider is configured without a custom base URL
- **THEN** it SHALL connect to `https://api.openai.com/v1`

#### Scenario: Custom base URL
- **WHEN** provider is configured with a `baseUrl` setting
- **THEN** it SHALL connect to that URL instead of the OpenAI default

### Requirement: Multi-Backend Support
The system SHALL support multiple services using the OpenAI-compatible API.

#### Scenario: Ollama backend
- **WHEN** provider base URL is set to `http://localhost:11434/v1`
- **THEN** it SHALL work with local Ollama models

#### Scenario: Groq backend
- **WHEN** provider base URL is set to `https://api.groq.com/openai/v1`
- **THEN** it SHALL work with Groq's fast inference

#### Scenario: Together AI backend
- **WHEN** provider base URL is set to `https://api.together.xyz/v1`
- **THEN** it SHALL work with Together AI's open-source models

### Requirement: Streaming Chat Completions
The system SHALL support streaming responses from OpenAI-compatible endpoints.

#### Scenario: Streaming enabled
- **WHEN** Generate is called
- **THEN** it SHALL use the streaming Chat Completions API
- **AND** SHALL yield StreamEvents as chunks arrive

### Requirement: Tool Calling Support
The system SHALL support function/tool calling via the OpenAI tools API.

#### Scenario: Tools provided
- **WHEN** GenerateParams contains Tools
- **THEN** they SHALL be converted to OpenAI tool format
- **AND** tool call responses SHALL be converted to StreamEvent format

### Requirement: API Key Configuration
The system SHALL support API key authentication.

#### Scenario: API key from config
- **WHEN** provider config contains `apiKey`
- **THEN** it SHALL be used in the Authorization header

#### Scenario: No API key required
- **WHEN** provider config has no `apiKey` and baseUrl is local (e.g., Ollama)
- **THEN** requests SHALL proceed without authentication

### Requirement: ListModels debug logging
The OpenAI provider's `ListModels()` method SHALL log debug messages for request start, success (with model count), and failure (with error).

#### Scenario: Successful model listing logged
- **WHEN** `ListModels()` succeeds and returns models
- **THEN** a debug log SHALL be emitted with provider ID and model count

#### Scenario: Failed model listing logged
- **WHEN** `ListModels()` fails with an error
- **THEN** a debug log SHALL be emitted with provider ID and error details

### Requirement: OpenAI provider forwards streaming tool call Index
The OpenAI provider SHALL forward the `Index` field from the SDK's `ToolCall` struct to the `provider.ToolCall.Index` field in streaming events.

#### Scenario: Streaming chunk with Index
- **WHEN** an OpenAI streaming delta contains a ToolCall with Index=0
- **THEN** the emitted `provider.StreamEvent.ToolCall.Index` is a pointer to 0

### Requirement: convertParams filters empty-name tools
The `convertParams` function SHALL exclude tools with empty Name from the request's Tools array.

#### Scenario: Tool with empty name
- **WHEN** a `GenerateParams` contains a Tool with Name=""
- **THEN** the converted request's Tools array does not include that tool

#### Scenario: All tools valid
- **WHEN** all Tools have non-empty Names
- **THEN** the converted request's Tools array contains all of them unchanged

### Requirement: convertParams filters empty-name tool calls in messages
The `convertParams` function SHALL exclude tool calls with empty Name from message ToolCalls arrays.

#### Scenario: ToolCall with empty name in message
- **WHEN** a message contains a ToolCall with Name=""
- **THEN** the converted message's ToolCalls array does not include that entry

#### Scenario: All tool calls valid
- **WHEN** all ToolCalls in a message have non-empty Names
- **THEN** the converted message's ToolCalls array contains all of them unchanged
