## ADDED Requirements

### Requirement: Empty response fallback for channel path
The `runAgent` function in `channels.go` SHALL return a user-visible fallback message when the agent succeeds (no error) but produces an empty response string. The fallback message SHALL be a package-level constant.

#### Scenario: Agent returns empty response via channel
- **WHEN** the agent `RunAndCollect` returns an empty string with no error
- **THEN** `runAgent` SHALL substitute the `emptyResponseFallback` constant as the response
- **AND** SHALL log a warning with session key and elapsed time

#### Scenario: Agent returns non-empty response via channel
- **WHEN** the agent `RunAndCollect` returns a non-empty string with no error
- **THEN** `runAgent` SHALL return the response unchanged

### Requirement: Empty response fallback for gateway path
The `handleChatMessage` function in `gateway/server.go` SHALL return a user-visible fallback message when the agent succeeds but produces an empty response string via the WebSocket streaming path.

#### Scenario: Agent returns empty response via gateway
- **WHEN** the agent `RunStreaming` returns an empty string with no error
- **THEN** `handleChatMessage` SHALL substitute the `emptyResponseFallback` constant as the response
- **AND** SHALL log a warning with session key

#### Scenario: Agent returns error via gateway
- **WHEN** the agent `RunStreaming` returns an error
- **THEN** `handleChatMessage` SHALL NOT apply the fallback and SHALL propagate the error normally

### Requirement: Agent empty response diagnostic logging
The `Agent.RunAndCollect` function SHALL log a warning when the agent run succeeds but produces an empty response string, providing session ID and elapsed time for diagnostics.

#### Scenario: Empty response logged at agent level
- **WHEN** `runAndCollectOnce` returns an empty string with no error
- **THEN** `RunAndCollect` SHALL log a warn-level message with session and elapsed fields

### Requirement: Agent text collection without thought filter
The `runAndCollectOnce` and `RunStreaming` functions SHALL collect text from session event parts using `part.Text != ""` without filtering on `part.Thought`. Thought filtering is the responsibility of the provider layer, not the agent layer.

#### Scenario: Text parts collected without thought check
- **WHEN** a session event contains text parts
- **THEN** the agent SHALL collect all non-empty text parts regardless of the `Thought` field value
