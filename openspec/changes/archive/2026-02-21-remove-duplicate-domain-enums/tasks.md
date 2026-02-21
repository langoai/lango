# Tasks

## 1. Remove duplicate domain enum types
- [x] 1.1 Remove `KnowledgeCategory` type, constants, `Valid()`, `Values()` from `internal/knowledge/types.go`
- [x] 1.2 Remove `LearningCategory` type, constants, `Valid()`, `Values()` from `internal/knowledge/types.go`
- [x] 1.3 Change `KnowledgeEntry.Category` field type to `entknowledge.Category`
- [x] 1.4 Change `LearningEntry.Category` field type to `entlearning.Category`

## 2. Remove boundary casts in store
- [x] 2.1 Remove `SetCategory(entknowledge.Category(string(entry.Category)))` → `SetCategory(entry.Category)` in `store.go`
- [x] 2.2 Remove `KnowledgeCategory(k.Category)` → `k.Category` in read paths
- [x] 2.3 Apply same pattern for Learning entries

## 3. Update learning package
- [x] 3.1 Change `mapKnowledgeCategory()` return type to `entknowledge.Category` in `parse.go`
- [x] 3.2 Change `mapLearningCategory()` return type to `entlearning.Category` in `parse.go`
- [x] 3.3 Change `categorizeError()` return type to `entlearning.Category` in `analyzer.go`
- [x] 3.4 Update `engine.go`, `session_learner.go`, `conversation_analyzer.go` to use `entlearning.CategoryUserCorrection`

## 4. Update librarian package
- [x] 4.1 Change `mapCategory()` return type to `entknowledge.Category` in `proactive_buffer.go`
- [x] 4.2 Update `entknowledge.Category()` cast in `inquiry_processor.go`

## 5. Update app tools
- [x] 5.1 Use `entknowledge.Category(category)` in `save_knowledge` handler
- [x] 5.2 Use `entlearning.Category(category)` in `save_learning` handler

## 6. Update tests
- [x] 6.1 Update `store_test.go` to use `entlearning.CategoryXxx` constants
- [x] 6.2 Update `retriever_test.go` to use `entlearning.CategoryXxx` constants
- [x] 6.3 Update `analyzer_test.go` to use `entlearning.CategoryXxx` constants
- [x] 6.4 Update `engine_test.go` to use `entlearning.CategoryXxx` constants

## 7. Verification
- [x] 7.1 `go build ./...` — zero errors
- [x] 7.2 `go test ./internal/knowledge/... ./internal/learning/...` — all pass
- [x] 7.3 Zero references to `knowledge.KnowledgeCategory` / `knowledge.LearningCategory` in Go files
