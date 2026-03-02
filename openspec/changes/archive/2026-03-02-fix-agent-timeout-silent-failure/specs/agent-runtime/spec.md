## ADDED Requirements

### Requirement: Context deadline detection after ADK iterator completion
The `runAndCollectOnce` and `RunStreaming` methods SHALL check `ctx.Err()` after the ADK iterator completes to detect context deadline exceeded errors that the ADK streaming iterator fails to propagate.

#### Scenario: Context deadline exceeded during iteration
- **WHEN** the context deadline expires while iterating over ADK runner events
- **AND** the ADK iterator terminates without yielding an error
- **THEN** `runAndCollectOnce` SHALL check `ctx.Err()` after the iteration loop
- **AND** SHALL return an error wrapping the context error if `ctx.Err()` is non-nil

#### Scenario: Context deadline exceeded during streaming
- **WHEN** the context deadline expires while `RunStreaming` iterates over ADK runner events
- **AND** the ADK iterator terminates without yielding an error
- **THEN** `RunStreaming` SHALL check `ctx.Err()` after the iteration loop
- **AND** SHALL return an error wrapping the context error if `ctx.Err()` is non-nil

#### Scenario: Normal completion without context error
- **WHEN** the ADK iterator completes normally
- **AND** `ctx.Err()` returns `nil`
- **THEN** the collected response text SHALL be returned without error
