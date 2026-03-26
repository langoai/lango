## 1. Category Infrastructure

- [x] 1.1 Create `internal/knowledge/category.go` with exported `MapKnowledgeCategory()` and `MapLearningCategory()` (error return)
- [x] 1.2 Create `internal/knowledge/category_test.go` with table-driven tests for all 12 valid inputs + error cases
- [x] 1.3 Fix `mapLearningCategory()` in `internal/learning/parse.go` to return `(Category, error)` instead of silent fallback

## 2. Analysis Struct Extensions

- [x] 2.1 Add `Temporal string` field to `analysisResult` struct in `internal/learning/parse.go`
- [x] 2.2 Add `Temporal string` field to `ObservationKnowledge` struct in `internal/librarian/types.go`
- [x] 2.3 Add `Temporal string` field to `matchedKnowledge` struct in `internal/librarian/parse.go`

## 3. Prompt Standardization

- [x] 3.1 Update ConversationAnalyzer prompt to all 6 types + temporal hint
- [x] 3.2 Update SessionLearner prompt to all 6 types + temporal hint
- [x] 3.3 Update ObservationAnalyzer prompt to include temporal hint
- [x] 3.4 Update InquiryProcessor prompt to all 6 types + temporal hint

## 4. Routing — Dual-Save + All-As-Knowledge

- [x] 4.1 Create shared `saveAnalysisResult()` in `parse.go` — all types → knowledge, pattern/correction → also learning
- [x] 4.2 Update ConversationAnalyzer.saveResult() to delegate to shared helper
- [x] 4.3 Update SessionLearner.saveSessionResult() to delegate to shared helper
- [x] 4.4 Add temporal tag handling in `saveAnalysisResult()` — append `"temporal:"+r.Temporal` to Tags
- [x] 4.5 Add dual-save in ProactiveBuffer — pattern/correction extractions also SaveLearning
- [x] 4.6 Add dual-save in InquiryProcessor — pattern/correction matched knowledge also SaveLearning
- [x] 4.7 Add temporal tag handling in ProactiveBuffer — append temporal tag to knowledge entry
- [x] 4.8 Add temporal tag handling in InquiryProcessor — append temporal tag to knowledge entry

## 5. Content-Dedup

- [x] 5.1 Add content-dedup check in `saveKnowledgeOnce` — same (category, content) = no-op
- [x] 5.2 Add `TestSaveKnowledge_ContentDedup` test — same content no-op, different content creates version, different category creates version

## 6. Downstream Artifacts

- [x] 6.1 Update `save_knowledge` tool description in `tools_meta.go`
- [x] 6.2 Update README.md Knowledge Store description

## 7. Build Verification

- [x] 7.1 Run `CGO_ENABLED=1 go build -tags fts5 ./...` — zero build errors
- [x] 7.2 Run `CGO_ENABLED=1 go test -tags fts5 ./internal/knowledge/ ./internal/learning/ ./internal/librarian/ ./internal/app/` — all tests pass
