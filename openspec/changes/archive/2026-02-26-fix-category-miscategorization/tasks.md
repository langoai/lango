## 1. Category Mapping Functions

- [x] 1.1 Change `mapCategory()` in `internal/librarian/proactive_buffer.go` to return `(entknowledge.Category, error)`, add `"pattern"` and `"correction"` cases, return error for unrecognized types
- [x] 1.2 Change `mapKnowledgeCategory()` in `internal/learning/parse.go` to return `(entknowledge.Category, error)`, add `"pattern"` and `"correction"` cases, return error for unrecognized types

## 2. Caller Updates

- [x] 2.1 Update `ProactiveBuffer.process()` in `proactive_buffer.go` to handle `mapCategory()` error: log warning and skip extraction on error
- [x] 2.2 Update `ConversationAnalyzer.saveResult()` in `conversation_analyzer.go` to handle `mapKnowledgeCategory()` error
- [x] 2.3 Update `SessionLearner.saveSessionResult()` in `session_learner.go` to handle `mapKnowledgeCategory()` error
- [x] 2.4 Replace raw `entknowledge.Category()` cast in `InquiryProcessor.ProcessAnswers()` with `mapCategory()` call, skip knowledge save on error while still resolving inquiry

## 3. Tool & Prompt Updates

- [x] 3.1 Add `"pattern"` and `"correction"` to `save_knowledge` tool enum in `internal/app/tools.go`, validate via `entknowledge.CategoryValidator()` before saving
- [x] 3.2 Update observation analyzer prompt in `internal/librarian/observation_analyzer.go` to include `pattern|correction` in type list

## 4. Tests

- [x] 4.1 Add table-driven test for `mapCategory()` in `internal/librarian/proactive_buffer_test.go` covering all 6 valid types and 3 invalid cases
- [x] 4.2 Add table-driven test for `mapKnowledgeCategory()` in `internal/learning/parse_test.go` covering all 6 valid types and 3 invalid cases

## 5. Verification

- [x] 5.1 Run `go build ./...` — confirm compilation succeeds
- [x] 5.2 Run `go test ./...` — confirm all tests pass
