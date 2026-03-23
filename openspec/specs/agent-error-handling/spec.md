# agent-error-handling Specification

## Purpose
Structured error types for agent execution failures with error classification, partial result preservation, and user-facing messages across all channels.
## Requirements
### Requirement: Structured agent error type
The system SHALL provide an `AgentError` type with fields: `Code` (ErrorCode), `Message` (string), `Cause` (error), `Partial` (string), and `Elapsed` (time.Duration). It SHALL implement the `error` and `Unwrap` interfaces.

#### Scenario: AgentError implements error interface
- **WHEN** an `AgentError` is created with Code `ErrTimeout` and Cause `context.DeadlineExceeded`
- **THEN** calling `Error()` SHALL return a string containing the error code and cause message

#### Scenario: AgentError supports errors.As unwrapping
- **WHEN** an `AgentError` is wrapped in `fmt.Errorf("outer: %w", agentErr)`
- **THEN** `errors.As(wrappedErr, &target)` SHALL succeed and populate the target with the original AgentError

### Requirement: Error classification
The system SHALL classify errors into codes: `ErrTimeout` (E001), `ErrModelError` (E002), `ErrToolError` (E003), `ErrTurnLimit` (E004), `ErrInternal` (E005), `ErrIdleTimeout` (E006), `ErrToolChurn` (E007), and `ErrEmptyAfterToolUse` (E008). Classification SHALL be based on error content and context state.

#### Scenario: Context deadline classified as timeout
- **WHEN** the error is or wraps `context.DeadlineExceeded`
- **THEN** `classifyError` SHALL return `ErrTimeout`

#### Scenario: Turn limit error classified correctly
- **WHEN** the error message contains "maximum turn limit"
- **THEN** `classifyError` SHALL return `ErrTurnLimit`

#### Scenario: Approval failure classified as tool error
- **WHEN** the error wraps `approval.ErrDenied`, `approval.ErrTimeout`, or `approval.ErrUnavailable`
- **THEN** `classifyError` SHALL return `ErrToolError`

#### Scenario: thought_signature error classified as model error
- **WHEN** the error message contains both `"function call"` and `"thought_signature"` (e.g., Gemini API error `"Function call is missing a thought_signature in functionCall parts"`)
- **THEN** `classifyError` SHALL return `ErrModelError`, not `ErrToolError`
- **AND** the `thought_signature` check SHALL be evaluated BEFORE the `"tool"`/`"function call"` keyword check

#### Scenario: Pure tool error still classified correctly
- **WHEN** the error message contains `"tool"` but not `"thought_signature"`
- **THEN** `classifyError` SHALL return `ErrToolError`

#### Scenario: Tool churn error classified correctly
- **WHEN** the error message contains "consecutively, forcing stop"
- **THEN** `classifyError` SHALL return `ErrToolChurn`

#### Scenario: Empty-after-tool-use classified correctly
- **WHEN** the runtime terminates a turn after successful tool activity but without any visible assistant completion
- **THEN** it SHALL return `ErrEmptyAfterToolUse`
- **AND** the turn SHALL NOT be treated as a successful empty response

#### Scenario: Unknown error classified as internal
- **WHEN** the error does not match any known pattern
- **THEN** `classifyError` SHALL return `ErrInternal`

### Requirement: User-facing error messages
The `AgentError` SHALL provide a `UserMessage()` method that returns a human-readable message including the error code and actionable guidance. User-facing messages SHALL NOT instruct the user to read a raw partial draft above.

#### Scenario: Timeout with partial result
- **WHEN** an `AgentError` has Code `ErrTimeout` and a non-empty `Partial` field
- **THEN** `UserMessage()` SHALL report the timeout with actionable guidance
- **AND** it SHALL NOT claim that the partial response was shown to the user

#### Scenario: Timeout without partial result
- **WHEN** an `AgentError` has Code `ErrTimeout` and an empty `Partial` field
- **THEN** `UserMessage()` SHALL suggest breaking the question into smaller parts

#### Scenario: Approval denied message
- **WHEN** the underlying error wraps `approval.ErrDenied`
- **THEN** `UserMessage()` SHALL explain that the action was denied by approval

#### Scenario: Approval expired message
- **WHEN** the underlying error wraps `approval.ErrTimeout`
- **THEN** `UserMessage()` SHALL explain that the approval request expired

#### Scenario: Approval unavailable message
- **WHEN** the underlying error wraps `approval.ErrUnavailable`
- **THEN** `UserMessage()` SHALL explain that no approval channel was available

### Requirement: Streaming partial events omit thought metadata
Partial tool-call `LLMResponse` events yielded during streaming SHALL NOT carry `Thought` or `ThoughtSignature` fields. Only the final accumulated response (via `toolAccum.done()`) SHALL include the correct `Thought` and `ThoughtSignature` values.

#### Scenario: Partial tool-call event has no thought fields
- **WHEN** `ModelAdapter.GenerateContent` yields a partial event for a tool call with `Name` set
- **THEN** the `genai.Part` SHALL have `Thought=false` and `ThoughtSignature=nil`

#### Scenario: Final accumulated event preserves thought fields
- **WHEN** `ModelAdapter.GenerateContent` yields the final `TurnComplete=true` event
- **THEN** the accumulated `genai.Part` from `toolAccum.done()` SHALL preserve the original `Thought` and `ThoughtSignature` values from the stream

### Requirement: Partial result preservation on agent error
When an agent run fails (timeout, turn limit, or other error), the system SHALL return the accumulated text as the `Partial` field of the `AgentError` instead of discarding it.

#### Scenario: Timeout preserves partial text
- **WHEN** the agent has accumulated text "Here is a partial..." and the context deadline fires
- **THEN** the returned `AgentError` SHALL have `Partial` equal to "Here is a partial..."

#### Scenario: Iterator error preserves partial text
- **WHEN** the agent iterator yields an error after producing some text chunks
- **THEN** the returned `AgentError` SHALL have `Partial` containing the accumulated chunks

### Requirement: Partial result recovery in runAgent
When `runAgent()` receives an `AgentError` with a non-empty `Partial`, it SHALL retain the partial internally for diagnostics but SHALL NOT return the raw partial text to the user.

#### Scenario: Partial result suppressed from user response
- **WHEN** the agent returns an `AgentError` with `Partial` "Here is my analysis..."
- **THEN** `runAgent()` SHALL return only a user-facing warning/error note, and `nil` error
- **AND** it SHALL NOT append the raw partial draft to that message

#### Scenario: Error without partial propagated normally
- **WHEN** the agent returns an `AgentError` with empty `Partial`
- **THEN** `runAgent()` SHALL return the error to the channel for error display

### Requirement: Channel error formatting
All channel `sendError()` functions SHALL use `formatChannelError()` which checks for a `UserMessage()` method via duck-typed interface assertion, falling back to `Error()` for plain errors.

#### Scenario: AgentError formatted with UserMessage
- **WHEN** a channel receives an error implementing `UserMessage()`
- **THEN** the displayed error SHALL use the `UserMessage()` output

#### Scenario: Plain error formatted with Error
- **WHEN** a channel receives a plain error without `UserMessage()`
- **THEN** the displayed error SHALL use `Error()` output prefixed with "Error:"

### Requirement: Empty-after-tool-use error classification
If an agent run terminates with no visible assistant completion after one or more successful tool results, the runtime SHALL return a dedicated structured agent error classification for that condition rather than reporting a silent success.

#### Scenario: Channel path receives classified empty-after-tool-use error
- **WHEN** `RunAndCollect` finishes without visible text after successful specialist tool results
- **THEN** the runtime SHALL return a structured agent error classified as `empty_after_tool_use`
- **AND** the channel path SHALL surface the corresponding user-facing message instead of relying on the generic empty-success fallback

#### Scenario: Gateway path broadcasts classified empty-after-tool-use error
- **WHEN** `RunStreaming` finishes without visible text after successful specialist tool results
- **THEN** the gateway path SHALL emit an `agent.error` event carrying the `empty_after_tool_use` classification
- **AND** SHALL NOT treat the turn as a successful empty response

### Requirement: Call-signature loop classification
The runtime SHALL classify repeated same-agent same-tool same-params sequences as loop failures even when call IDs differ or tool response events are interleaved between attempts.

#### Scenario: Different call IDs do not bypass loop classification
- **WHEN** the same agent repeatedly emits the same tool name with canonically equal params but different generated call IDs
- **THEN** the runtime SHALL still classify the sequence as the same loop signature
- **AND** SHALL terminate the run with `loop_detected` once the configured threshold is exceeded

#### Scenario: Tool response interleaving does not reset the loop signature
- **WHEN** a specialist alternates `FunctionCall` and `FunctionResponse` events for the same canonical tool signature without visible assistant progress
- **THEN** the runtime SHALL continue counting the repeated signature
- **AND** SHALL NOT reset the loop counter solely because tool responses were observed

### Requirement: Truthful recovery messaging
User-facing recovery messages after loop, timeout, or empty-after-tool-use failures SHALL only describe evidence actually gathered during the current turn. They SHALL NOT claim that unavailable tools were directly executed.

#### Scenario: Recovery message avoids unavailable direct-call claim
- **WHEN** the orchestrator has no direct access to `payment_balance` and recovery messaging is generated after a specialist failure
- **THEN** the user-facing message SHALL NOT claim that the orchestrator directly executed `payment_balance`
- **AND** SHALL instead describe the actual specialist failure or previously gathered evidence truthfully
