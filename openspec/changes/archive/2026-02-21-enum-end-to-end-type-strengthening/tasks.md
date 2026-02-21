## 1. Phase 1: Message.Role → types.MessageRole

- [x] 1.1 Change `session.Message.Role` from `string` to `types.MessageRole` in `internal/session/store.go`
- [x] 1.2 Add `string()` cast at Ent write boundaries in `internal/session/ent_store.go` (`SetRole(string(msg.Role))`)
- [x] 1.3 Add `types.MessageRole()` cast at Ent read boundary in `internal/session/ent_store.go`
- [x] 1.4 Update `internal/adk/session_service.go` to use `types.MessageRole.Normalize()` directly and typed constants (`types.RoleAssistant`, `types.RoleUser`)
- [x] 1.5 Update `internal/adk/state.go` switch cases to use typed constants and add `string(role)` cast at genai boundary

## 2. Phase 2: parseDeliveryTarget → types.ChannelType

- [x] 2.1 Change `parseDeliveryTarget()` return type from `(string, string)` to `(types.ChannelType, string)` in `internal/app/sender.go`
- [x] 2.2 Remove `string(types.ChannelXXX)` casts in `SendMessage` and `StartTyping` switch statements

## 3. Phase 3: Confidence → types.Confidence

- [x] 3.1 Change `analysisResult.Confidence` from `string` to `types.Confidence` in `internal/learning/parse.go`
- [x] 3.2 Remove `string()` cast in `internal/learning/session_learner.go` confidence comparison
- [x] 3.3 Change `ObservationKnowledge.Confidence` from `string` to `types.Confidence` in `internal/librarian/types.go`
- [x] 3.4 Change `answerMatch.Confidence` from `string` to `types.Confidence` in `internal/librarian/parse.go`
- [x] 3.5 Remove `string()` cast in `internal/librarian/inquiry_processor.go` confidence comparison
- [x] 3.6 Change `ProactiveBuffer.autoSaveConfidence` and `ProactiveBufferConfig.AutoSaveConfidence` to `types.Confidence`, remove all `string()` casts in `shouldAutoSave()`
- [x] 3.7 Change `LibrarianConfig.AutoSaveConfidence` to `types.Confidence` in `internal/config/types.go`
- [x] 3.8 Update config default in `internal/config/loader.go` to use `types.ConfidenceHigh`
- [x] 3.9 Add boundary casts in CLI: `types.Confidence(val)` in `state_update.go`, `string(cfg...)` in `forms_impl.go`

## 4. Phase 4: KnowledgeEntry.Category → knowledge.KnowledgeCategory

- [x] 4.1 Change `KnowledgeEntry.Category` from `string` to `KnowledgeCategory` in `internal/knowledge/types.go`
- [x] 4.2 Add `string()` casts at Ent write boundaries and `KnowledgeCategory()` at read boundaries in `internal/knowledge/store.go`
- [x] 4.3 Add `string()` cast in `internal/knowledge/retriever.go` when populating `ContextItem.Category`
- [x] 4.4 Change `mapKnowledgeCategory()` return type to `knowledge.KnowledgeCategory` in `internal/learning/parse.go`
- [x] 4.5 Change `mapCategory()` return type to `knowledge.KnowledgeCategory` in `internal/librarian/proactive_buffer.go`
- [x] 4.6 Add `knowledge.KnowledgeCategory()` boundary cast in `save_knowledge` tool handler in `internal/app/tools.go`
- [x] 4.7 Add `knowledge.KnowledgeCategory()` boundary cast in `internal/librarian/inquiry_processor.go`

## 5. Verification

- [x] 5.1 `go build ./...` passes with zero errors
- [x] 5.2 `go test ./...` passes with all tests passing
- [x] 5.3 Verify `string(types.XXX)` casts only remain in CLI packages (valid system boundary)
