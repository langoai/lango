## Requirements

### Requirement: Multi-Provider Support
The agent runtime SHALL delegate all LLM interactions to the ADK agent (`internal/adk/agent.go`). The legacy `Runtime.Run()` execution loop SHALL be removed. Type definitions (`Tool`, `ToolHandler`, `StreamEvent`, `Config`, `ParameterDef`, `AdkToolAdapter`) SHALL be retained for use by tool registration code.

#### Scenario: Provider initialization
- **WHEN** the application creates an ADK agent
- **THEN** it SHALL use `supervisor.NewProviderProxy()` and `adk.NewModelAdapter()` to bridge the provider to ADK
- **THEN** the `adk.Agent.Run()` method SHALL be the only execution path for generating responses

#### Scenario: Legacy runtime removal
- **WHEN** code references `agent.Runtime`
- **THEN** only type definitions and tool registration methods (`RegisterTool`, `GetTool`, `ListTools`, `ExecuteTool`) SHALL be available
- **THEN** the `Run()` method SHALL NOT exist

### Requirement: Model Fallback execution
The system SHALL execute model fallbacks when the primary provider fails.

#### Scenario: Primary provider failure
- **WHEN** the primary provider fails with a retryable error (e.g., rate limit, overload)
- **THEN** the system SHALL attempt to use the next configured fallback provider
- **AND** SHALL log the fallback event

#### Scenario: All providers fail
- **WHEN** all configured providers (primary + fallbacks) fail
- **THEN** the system SHALL return an error to the user

### Requirement: Agent Run Timing and Result Logging
The ADK agent `RunAndCollect` method SHALL log timing and result information for every execution attempt.

#### Scenario: Successful first attempt
- **WHEN** `RunAndCollect` completes successfully on the first attempt
- **THEN** the system SHALL log at Debug level: session ID, elapsed time, response length

#### Scenario: Non-hallucination failure
- **WHEN** `RunAndCollect` fails with a non-hallucination error
- **THEN** the system SHALL log at Warn level: session ID, elapsed time, error message

#### Scenario: Hallucination retry success
- **WHEN** the first attempt fails with a hallucination error and the retry succeeds
- **THEN** the system SHALL log at Info level: session ID, retry elapsed time, total elapsed time, response length

#### Scenario: Hallucination retry failure
- **WHEN** the first attempt fails with a hallucination error and the retry also fails
- **THEN** the system SHALL log at Error level: session ID, retry elapsed time, total elapsed time, error message

### Requirement: Channel Request Lifecycle Logging
The `runAgent` method SHALL log timing information for the full request lifecycle.

#### Scenario: Request started
- **WHEN** `runAgent` begins processing a request
- **THEN** the system SHALL log at Debug level: session key, configured timeout

#### Scenario: Request completed
- **WHEN** `runAgent` completes successfully
- **THEN** the system SHALL log at Info level: session key, elapsed time, response length

#### Scenario: Request failed
- **WHEN** `runAgent` fails with a non-timeout error
- **THEN** the system SHALL log at Warn level: session key, elapsed time, error message

#### Scenario: Request timed out
- **WHEN** `runAgent` fails due to context deadline exceeded
- **THEN** the system SHALL log at Error level: session key, elapsed time, configured timeout

#### Scenario: Approaching timeout warning
- **WHEN** 80% of the configured timeout has elapsed and the request is still running
- **THEN** the system SHALL log at Warn level: session key, elapsed time so far, configured timeout
- **AND** the warning timer SHALL be cancelled if the request completes before 80%

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
