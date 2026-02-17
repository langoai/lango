## 1. Infrastructure

- [x] 1.1 Create Ent schema `internal/ent/schema/inquiry.go` with UUID id, session_key, topic, question, context, priority enum, status enum, answer, knowledge_key, source_observation_id, created_at, resolved_at fields and indexes
- [x] 1.2 Run `go generate ./internal/ent/...` to generate Ent code
- [x] 1.3 Add `LibrarianConfig` struct to `internal/config/types.go` with Enabled, ObservationThreshold, InquiryCooldownTurns, MaxPendingInquiries, AutoSaveConfidence, Provider, Model fields
- [x] 1.4 Add `Librarian LibrarianConfig` field to the root `Config` struct

## 2. Librarian Package - Domain Types

- [x] 2.1 Create `internal/librarian/types.go` with ObservationKnowledge, KnowledgeGap, AnalysisOutput, Inquiry, TextGenerator, GraphCallback, Triple types
- [x] 2.2 Create `internal/librarian/parse.go` with parseAnalysisOutput, parseAnswerMatches, stripCodeFence helpers

## 3. Librarian Package - Core Components

- [x] 3.1 Create `internal/librarian/inquiry_store.go` with InquiryStore providing SaveInquiry, ListPendingInquiries, ResolveInquiry, DismissInquiry, CountPendingBySession methods
- [x] 3.2 Create `internal/librarian/observation_analyzer.go` with ObservationAnalyzer using TextGenerator for LLM-based observation analysis
- [x] 3.3 Create `internal/librarian/inquiry_processor.go` with InquiryProcessor for answer detection and knowledge saving
- [x] 3.4 Create `internal/librarian/proactive_buffer.go` with ProactiveBuffer implementing Start/Trigger/Stop lifecycle with two-phase processing

## 4. Knowledge Context Layer Extension

- [x] 4.1 Add `LayerPendingInquiries` constant to `internal/knowledge/types.go`
- [x] 4.2 Add `InquiryProvider` interface to `internal/knowledge/types.go`
- [x] 4.3 Add `inquiryProvider` field and `WithInquiryProvider()` builder method to `internal/knowledge/retriever.go`
- [x] 4.4 Add `LayerPendingInquiries` case in ContextRetriever.Retrieve()
- [x] 4.5 Add "Pending Knowledge Inquiries" section in ContextRetriever.AssemblePrompt()

## 5. Agent Integration

- [x] 5.1 Update librarian AgentSpec in `internal/orchestration/tools.go` with `librarian_` prefix, inquiry/question/gap keywords, and proactive behavior instruction
- [x] 5.2 Add `librarian_` to capabilityMap in `internal/orchestration/tools.go`
- [x] 5.3 Add `buildLibrarianTools()` to `internal/app/tools.go` with librarian_pending_inquiries and librarian_dismiss_inquiry tools
- [x] 5.4 Add `inquiryProviderAdapter` struct to `internal/app/wiring.go` bridging InquiryStore to InquiryProvider
- [x] 5.5 Add `librarianComponents` struct and `initLibrarian()` function to `internal/app/wiring.go`
- [x] 5.6 Add LibrarianInquiryStore and LibrarianProactiveBuffer fields to `internal/app/types.go`

## 6. App Wiring

- [x] 6.1 Call `initLibrarian()` in `internal/app/app.go` after conversation analysis init
- [x] 6.2 Append librarian tools to tool chain in `internal/app/app.go`
- [x] 6.3 Wire InquiryProvider into ContextRetriever via initAgent parameter
- [x] 6.4 Register ProactiveBuffer.Trigger in gateway OnTurnComplete callback
- [x] 6.5 Add ProactiveBuffer.Start to App.Start() and ProactiveBuffer.Stop to App.Stop()

## 7. Verification

- [x] 7.1 Run `go build ./...` and verify no build errors
- [x] 7.2 Run `go test ./internal/librarian/...` for new package tests
- [x] 7.3 Run `go test ./internal/knowledge/...` for retriever changes
- [x] 7.4 Run `go test ./internal/orchestration/...` for tool partition changes
- [x] 7.5 Run `go test ./...` for full test suite
