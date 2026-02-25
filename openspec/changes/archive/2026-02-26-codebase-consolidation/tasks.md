# Tasks: Codebase Consolidation

## Phase 1: Generic AsyncBuffer
- [x] 1.1 Create `internal/asyncbuf/batch.go` — BatchBuffer[T] generic type
- [x] 1.2 Create `internal/asyncbuf/trigger.go` — TriggerBuffer[T] generic type
- [x] 1.3 Create `internal/asyncbuf/batch_test.go` — 6 tests for BatchBuffer
- [x] 1.4 Create `internal/asyncbuf/trigger_test.go` — 5 tests for TriggerBuffer
- [x] 1.5 Migrate `internal/embedding/buffer.go` to wrap BatchBuffer[EmbedRequest]
- [x] 1.6 Migrate `internal/graph/buffer.go` to wrap BatchBuffer[GraphRequest]
- [x] 1.7 Migrate `internal/memory/buffer.go` to wrap TriggerBuffer[string]
- [x] 1.8 Migrate `internal/learning/analysis_buffer.go` to wrap TriggerBuffer[AnalysisRequest]
- [x] 1.9 Migrate `internal/librarian/proactive_buffer.go` to wrap TriggerBuffer[string]

## Phase 2: Package Merges
- [x] 2.1 Move ctxutil/Detach to `internal/types/context.go`, update importers, delete ctxutil/
- [x] 2.2 Move passphrase/ to `internal/security/passphrase/`, update importers, delete passphrase/
- [x] 2.3 Move zkp/ to `internal/p2p/zkp/`, update importers, delete zkp/

## Phase 3: CLI UX
- [x] 3.1 Add Cobra command groups (core, config, data, infra) to root command
- [x] 3.2 Set GroupID on all commands
- [x] 3.3 Add "See Also" cross-references to config, settings, onboard, doctor

## Phase 4: Type Consolidation (SKIPPED)
- [x] 4.1 ~~Consolidate MessageProvider type~~ — SKIPPED: types→session→types import cycle
