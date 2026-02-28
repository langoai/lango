## ADDED Requirements

### Requirement: Anthropic Messages API
The system SHALL implement a provider using the Anthropic Messages API.

#### Scenario: API endpoint
- **WHEN** Anthropic provider is used
- **THEN** it SHALL connect to `https://api.anthropic.com/v1/messages`

### Requirement: Streaming Messages
The system SHALL support streaming responses from the Anthropic API.

#### Scenario: Stream delta events
- **WHEN** Generate is called
- **THEN** it SHALL use the streaming Messages API
- **AND** SHALL yield StreamEvents for `content_block_delta` events

### Requirement: Claude Model Support
The system SHALL support Claude model families.

#### Scenario: Claude 3.5 models
- **WHEN** model is `claude-3-5-sonnet-20241022` or similar
- **THEN** it SHALL work correctly

#### Scenario: Claude 3 models
- **WHEN** model is `claude-3-opus`, `claude-3-sonnet`, or `claude-3-haiku`
- **THEN** it SHALL work correctly

### Requirement: Tool Use Support
The system SHALL support Anthropic's tool use format.

#### Scenario: Tool definition
- **WHEN** GenerateParams contains Tools
- **THEN** they SHALL be converted to Anthropic's `tools` format with `input_schema`

#### Scenario: Tool use response
- **WHEN** Claude responds with `tool_use` content block
- **THEN** it SHALL be converted to a StreamEvent with tool call details

### Requirement: Extended Thinking Support
The system SHALL support Claude's extended thinking feature when available.

#### Scenario: Thinking enabled
- **WHEN** model supports extended thinking and it is requested
- **THEN** reasoning content SHALL be included in the response metadata

### Requirement: Anthropic provider unknown role handling
The `convertParams` method SHALL handle unknown message roles by logging a warning and skipping the message. The switch statement SHALL include explicit cases for "user", "assistant", and "system" roles, and a `default` case that logs the unknown role via the subsystem logger.

#### Scenario: Unknown role is logged and skipped
- **WHEN** `convertParams` encounters a message with role "tool" or any unrecognized role
- **THEN** it SHALL log a warning containing the unknown role value
- **AND** it SHALL NOT include that message in the Anthropic API request
- **AND** it SHALL NOT return an error

#### Scenario: System role handled separately
- **WHEN** `convertParams` encounters a message with role "system"
- **THEN** it SHALL NOT log a warning (system is handled in a separate loop)

### Requirement: Anthropic provider constructor accepts explicit ID
The Anthropic provider constructor SHALL accept an `id` string parameter and use it as the provider's registry identity, instead of hardcoding `"anthropic"`.

#### Scenario: Custom ID registration
- **WHEN** `NewProvider("my-claude", "sk-ant-xxx")` is called
- **THEN** the returned provider's `ID()` method SHALL return `"my-claude"`

#### Scenario: Default ID registration
- **WHEN** `NewProvider("anthropic", "sk-ant-xxx")` is called
- **THEN** the returned provider's `ID()` method SHALL return `"anthropic"`

### Requirement: Live model listing
The Anthropic provider's `ListModels()` MUST call the Anthropic Models API instead of returning hardcoded values.

#### Scenario: Successful model listing
- **WHEN** ListModels is called with valid API credentials
- **THEN** returns all models from the API using paginated auto-paging with limit 1000

#### Scenario: Partial failure
- **WHEN** API returns some models before encountering an error
- **THEN** returns the successfully fetched models without error

#### Scenario: Complete failure
- **WHEN** API call fails with no models retrieved
- **THEN** returns error with wrapped context
