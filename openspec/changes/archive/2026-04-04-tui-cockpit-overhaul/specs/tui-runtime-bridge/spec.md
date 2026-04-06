## ADDED Requirements

### Requirement: Chat-internal bridge function
The chat package SHALL provide an `enrichRequest(program *tea.Program, req *turnrunner.Request)` function that wires turnrunner callbacks to Bubble Tea messages without overwriting existing callbacks.

#### Scenario: Existing callbacks preserved
- **WHEN** `enrichRequest()` is called on a Request that already has `OnChunk` set
- **THEN** the existing `OnChunk` callback is NOT overwritten

#### Scenario: OnToolCall wired to ToolStartedMsg
- **WHEN** `enrichRequest()` sets `req.OnToolCall` and the callback is invoked
- **THEN** a `ToolStartedMsg` is sent to the tea.Program

#### Scenario: OnToolResult wired to ToolFinishedMsg
- **WHEN** `enrichRequest()` sets `req.OnToolResult` and the callback is invoked
- **THEN** a `ToolFinishedMsg` is sent to the tea.Program with success, duration, and output preview

#### Scenario: OnThinking wired to ThinkingStartedMsg/ThinkingFinishedMsg
- **WHEN** `enrichRequest()` sets `req.OnThinking` and the callback is invoked with `started: true`
- **THEN** a `ThinkingStartedMsg` is sent to the tea.Program

### Requirement: Bridge does not create a separate package
The bridge function SHALL reside in `internal/cli/chat/bridge.go` as a package-internal function, not in a shared UIEvent package.

#### Scenario: No uievent package exists
- **WHEN** the codebase is searched for `internal/cli/uievent/`
- **THEN** no such package exists

### Requirement: turnrunner.Request OnToolCall callback
The `turnrunner.Request` struct SHALL include an `OnToolCall func(callID, toolName string, params map[string]any)` field called when a tool invocation begins.

#### Scenario: OnToolCall fired on EventToolCall
- **WHEN** `recordEvent()` processes a `part.FunctionCall`
- **THEN** `req.OnToolCall` is invoked with callID, tool name, and parameters

#### Scenario: Nil OnToolCall is safe
- **WHEN** `req.OnToolCall` is nil and a tool call event occurs
- **THEN** the event is silently skipped without panic

### Requirement: turnrunner.Request OnToolResult callback
The `turnrunner.Request` struct SHALL include an `OnToolResult func(callID, toolName string, success bool, duration time.Duration, preview string)` field called when a tool invocation completes.

#### Scenario: OnToolResult fired on EventToolResult
- **WHEN** `recordEvent()` processes a `part.FunctionResponse`
- **THEN** `req.OnToolResult` is invoked with callID, tool name, success status, computed duration, and output preview

#### Scenario: Duration computed from callID-startedAt map
- **WHEN** a tool result arrives for a callID that was recorded in the startedAt map
- **THEN** duration equals `time.Since(startedAt)` and the map entry is cleaned up

### Requirement: turnrunner.Request OnThinking callback
The `turnrunner.Request` struct SHALL include an `OnThinking func(agentName string, started bool, summary string)` field called when thinking is detected via `genai.Part.Thought`.

#### Scenario: OnThinking fired on thought boundary
- **WHEN** `recordEvent()` encounters the first `part.Thought == true` transition (not already in thinking state)
- **THEN** `req.OnThinking` is invoked with `started: true` and the thought text; subsequent thought chunks accumulate without re-firing start

#### Scenario: OnThinking finished with accumulated summary
- **WHEN** a non-thought part arrives after one or more thought chunks
- **THEN** `req.OnThinking` is invoked with `started: false` and the full accumulated thought text as summary
