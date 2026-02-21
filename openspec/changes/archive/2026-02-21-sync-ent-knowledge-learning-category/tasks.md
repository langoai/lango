## 1. Ent Knowledge Schema Sync

- [x] 1.1 Add `pattern`, `correction` to `field.Enum("category").Values(...)` in `internal/ent/schema/knowledge.go`
- [x] 1.2 Run `go generate ./internal/ent` to regenerate Ent code

## 2. LearningCategory Typed Enum

- [x] 2.1 Add `LearningCategory` type with 6 constants (`LearningToolError`, `LearningProviderError`, `LearningUserCorrection`, `LearningTimeout`, `LearningPermission`, `LearningGeneral`) and `Valid()`/`Values()` methods in `internal/knowledge/types.go`
- [x] 2.2 Change `LearningEntry.Category` from `string` to `LearningCategory` in `internal/knowledge/types.go`

## 3. Store Boundary Casts

- [x] 3.1 Add `string()` cast at Ent write boundary in `internal/knowledge/store.go` (`SetCategory(entlearning.Category(string(entry.Category)))`)
- [x] 3.2 Add `LearningCategory()` cast at Ent read boundaries in `internal/knowledge/store.go` (GetLearning, SearchLearnings)
- [x] 3.3 Add `string()` cast at ContextItem boundary in `internal/knowledge/retriever.go`

## 4. Internal Code Typed Enum Usage

- [x] 4.1 Change `mapLearningCategory()` return type to `knowledge.LearningCategory` in `internal/learning/parse.go`
- [x] 4.2 Change `categorizeError()` return type to `knowledge.LearningCategory` in `internal/learning/analyzer.go`
- [x] 4.3 Replace `"user_correction"` with `knowledge.LearningUserCorrection` in `internal/learning/engine.go`
- [x] 4.4 Replace `"user_correction"` with `knowledge.LearningUserCorrection` in `internal/learning/session_learner.go` and `conversation_analyzer.go`
- [x] 4.5 Add `knowledge.LearningCategory(category)` boundary cast in `save_learning` tool handler in `internal/app/tools.go`

## 5. Test Updates

- [x] 5.1 Update `internal/knowledge/store_test.go` LearningEntry Category string literals to typed constants
- [x] 5.2 Update `internal/knowledge/retriever_test.go` LearningEntry Category to typed constant
- [x] 5.3 Update `internal/learning/engine_test.go` Category values to `knowledge.LearningCategory` constants
- [x] 5.4 Update `internal/learning/analyzer_test.go` test table `want` type to `knowledge.LearningCategory`

## 6. Verification

- [x] 6.1 `go build ./...` passes with zero errors
- [x] 6.2 `go test ./internal/knowledge/... ./internal/learning/... ./internal/app/...` passes
