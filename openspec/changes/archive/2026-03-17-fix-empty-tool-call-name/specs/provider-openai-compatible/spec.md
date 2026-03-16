## MODIFIED Requirements

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
