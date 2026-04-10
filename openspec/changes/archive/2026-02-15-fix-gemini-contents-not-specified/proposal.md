## Why

When receiving Telegram messages in Docker, the Gemini API returns a `"GenerateContentRequest.contents: contents is not specified"` error. When the ADK runner reads events from the session, the user message just added is missing from the in-memory history, causing empty contents to be sent to the Gemini API. Additionally, the system prompt is not passed to the provider, causing agent personality/instructions to be ignored.

## What Changes

- **AppendEvent in-memory history sync**: `SessionServiceAdapter.AppendEvent` updates `SessionAdapter.sess.History` after DB save so the ADK runner can read the current turn's user message
- **SystemInstruction forwarding**: `ModelAdapter.GenerateContent` converts `req.Config.SystemInstruction` to a system message and passes it to the provider
- Related tests added

## Capabilities

### New Capabilities

### Modified Capabilities
- `adk-architecture`: Added in-memory history sync in AppendEvent, added SystemInstruction forwarding in ModelAdapter

## Impact

- `internal/adk/session_service.go`: AppendEvent method modified
- `internal/adk/model.go`: GenerateContent method modified, extractSystemText helper added
- `internal/adk/session_service_test.go`: New test file
- `internal/adk/model_test.go`: SystemInstruction test added
- `internal/adk/state_test.go`: mockStore modified (DB-only behavior simulation)
