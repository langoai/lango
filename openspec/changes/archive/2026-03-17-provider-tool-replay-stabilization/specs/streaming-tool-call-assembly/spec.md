## MODIFIED Requirements

### Requirement: Shared convertMessages does not perform orphan repair
The `convertMessages()` function SHALL NOT inject synthetic tool responses for orphaned FunctionCalls. Orphan repair is provider-specific and SHALL be handled by each provider's own conversion logic.

#### Scenario: Orphaned FunctionCall passes through without repair
- **WHEN** a genai.Content sequence contains a model FunctionCall followed by a user message with no intervening FunctionResponse
- **THEN** `convertMessages()` SHALL return the messages as-is without injecting synthetic tool messages
