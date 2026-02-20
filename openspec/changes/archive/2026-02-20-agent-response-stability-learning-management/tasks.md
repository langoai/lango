## 1. Rate Limit Removal

- [x] 1.1 Remove `maxKnowledge`, `maxLearnings` fields, `mu sync.Mutex`, `knowledgeCounts`, `learningCounts` from `Store` struct in `internal/knowledge/store.go`
- [x] 1.2 Remove `reserveKnowledgeSlot()` and `reserveLearningSlot()` functions from `internal/knowledge/store.go`
- [x] 1.3 Remove `reserveXxxSlot()` calls from `SaveKnowledge()` and `SaveLearning()`
- [x] 1.4 Change `NewStore()` signature from `(client, logger, maxKnowledge, maxLearnings)` to `(client, logger)`
- [x] 1.5 Remove `MaxLearnings` and `MaxKnowledge` from `KnowledgeConfig` in `internal/config/types.go`
- [x] 1.6 Remove defaults for `MaxLearnings` and `MaxKnowledge` from `internal/config/loader.go`
- [x] 1.7 Update `initKnowledge` in `internal/app/wiring.go` to use new `NewStore()` signature
- [x] 1.8 Remove `Max Learnings` and `Max Knowledge` form fields from `NewKnowledgeForm` in `internal/cli/settings/forms_impl.go`
- [x] 1.9 Remove `knowledge_max_learnings` and `knowledge_max_knowledge` cases from `internal/cli/tuicore/state_update.go`
- [x] 1.10 Update all test files using `NewStore` with old signature: `store_test.go`, `retriever_test.go`, `engine_test.go`, `analysis_buffer_test.go`, `session_learner_test.go`, `conversation_analyzer_test.go`
- [x] 1.11 Update `forms_impl_test.go` to remove references to removed fields

## 2. Learning Data Management Store Methods

- [x] 2.1 Add `LearningStats` type to `internal/knowledge/store.go`
- [x] 2.2 Implement `GetLearningStats()` method with Ent aggregation queries
- [x] 2.3 Implement `ListLearnings()` with category, minConfidence, olderThan filters and pagination
- [x] 2.4 Implement `DeleteLearning()` for single entry deletion by UUID
- [x] 2.5 Implement `DeleteLearningsWhere()` for bulk deletion with AND criteria
- [x] 2.6 Add tests for `GetLearningStats`, `ListLearnings`, `DeleteLearning`, `DeleteLearningsWhere` in `store_test.go`

## 3. Learning Data Management Agent Tools

- [x] 3.1 Add `learning_stats` tool to `buildMetaTools` in `internal/app/tools.go` (safety level: safe)
- [x] 3.2 Add `learning_cleanup` tool to `buildMetaTools` with dry_run support (safety level: moderate)

## 4. Agent Execution Observability

- [x] 4.1 Add timing logs to `RunAndCollect` in `internal/adk/agent.go`: success, failure, retry success, retry failure
- [x] 4.2 Add request lifecycle logs to `runAgent` in `internal/app/channels.go`: started, completed, failed, timed out
- [x] 4.3 Add 80% timeout approaching warning via `time.AfterFunc` in `runAgent`

## 5. Learning Engine Log Enhancement

- [x] 5.1 Add session key and tool name context to `handleError` save failure log in `internal/learning/engine.go`

## 6. Verification

- [x] 6.1 `go build ./...` passes
- [x] 6.2 `go test ./internal/knowledge/...` passes
- [x] 6.3 `go test ./internal/learning/...` passes
- [x] 6.4 `go test ./internal/adk/...` passes
- [x] 6.5 `go test ./internal/cli/settings/...` passes
- [x] 6.6 `go test ./internal/app/...` passes
