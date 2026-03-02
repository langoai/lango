## 1. Dependency Upgrade

- [x] 1.1 Upgrade ADK Go from v0.4.0 to v0.5.0 and run `go mod tidy`

## 2. ToolCall Struct Changes

- [x] 2.1 Add `Thought bool` and `ThoughtSignature []byte` to `provider.ToolCall` in `internal/provider/provider.go`
- [x] 2.2 Add `Thought bool` and `ThoughtSignature []byte` to `session.ToolCall` in `internal/session/store.go` with `omitempty` JSON tags
- [x] 2.3 Add `Thought bool` and `ThoughtSignature []byte` to `entschema.ToolCall` in `internal/ent/schema/message.go` with `omitempty` JSON tags

## 3. Gemini Provider Propagation

- [x] 3.1 Capture `part.Thought` and `part.ThoughtSignature` into `provider.ToolCall` in streaming response handler (`internal/provider/gemini/gemini.go`)
- [x] 3.2 Restore `Thought` and `ThoughtSignature` on `genai.Part` from `provider.ToolCall` in message builder (`internal/provider/gemini/gemini.go`)

## 4. ModelAdapter Propagation

- [x] 4.1 Set `Thought` and `ThoughtSignature` on `genai.Part` in streaming path (`internal/adk/model.go`)
- [x] 4.2 Set `Thought` and `ThoughtSignature` on `genai.Part` in non-streaming path (`internal/adk/model.go`)
- [x] 4.3 Extract `Thought` and `ThoughtSignature` from `genai.Part` into `provider.ToolCall` in `convertMessages` (`internal/adk/model.go`)

## 5. Session Persistence

- [x] 5.1 Capture `p.Thought` and `p.ThoughtSignature` into `internal.ToolCall` in `AppendEvent` (`internal/adk/session_service.go`)
- [x] 5.2 Restore `Thought` and `ThoughtSignature` on `genai.Part` in `EventsAdapter.All()` (`internal/adk/state.go`)

## 6. EntStore Conversion Points

- [x] 6.1 Propagate `Thought`/`ThoughtSignature` in `Create` method (`internal/session/ent_store.go`)
- [x] 6.2 Propagate `Thought`/`ThoughtSignature` in `AppendMessage` method (`internal/session/ent_store.go`)
- [x] 6.3 Propagate `Thought`/`ThoughtSignature` in `entToSession` method (`internal/session/ent_store.go`)

## 7. Verification

- [x] 7.1 `go build ./...` passes
- [x] 7.2 `go test ./...` passes
