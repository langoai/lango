## 1. Core Types & Data Model

- [x] 1.1 Create `internal/runledger/types.go` — RunStatus, StepStatus, ValidatorType, Step, Evidence, AcceptanceCriterion, PlannerOutput
- [x] 1.2 Create `internal/runledger/journal.go` — JournalEvent, JournalEventType, 13 typed payloads
- [x] 1.3 Create `internal/runledger/snapshot.go` — RunSnapshot, MaterializeFromJournal, ApplyTail, applyEvent, applyPolicyToSnapshot
- [x] 1.4 Create `internal/runledger/policy.go` — PolicyRequest, PolicyDecision, PolicyAction (7 types)
- [x] 1.5 Create `internal/runledger/errors.go` — Sentinel errors (ErrRunNotFound, ErrRunNotPaused, ErrStepNotFound, ErrAccessDenied, ErrRunCompleted)

## 2. Store & Caching

- [x] 2.1 Create `internal/runledger/store.go` — RunLedgerStore interface
- [x] 2.2 Implement MemoryStore with AppendJournalEvent, GetJournalEvents, GetJournalEventsSince
- [x] 2.3 Implement snapshot caching (GetCachedSnapshot + UpdateCachedSnapshot + tail-replay in GetRunSnapshot)
- [x] 2.4 Implement ListRuns with sorting and limit

## 3. Plan Parsing & Validation

- [x] 3.1 Create `internal/runledger/planparse.go` — ParsePlannerOutput (fenced JSON + raw JSON extraction)
- [x] 3.2 Implement ValidatePlanSchema — goal presence, step ID uniqueness, valid agents, valid validator types
- [x] 3.3 Implement dependency cycle detection via Kahn's algorithm
- [x] 3.4 Implement ConvertPlanToRunData with tool profile auto-inference from validator type

## 4. PEV Engine & Validators

- [x] 4.1 Create `internal/runledger/pev.go` — PEVEngine, Validator interface, Verify, VerifyAcceptanceCriteria
- [x] 4.2 Implement BuildPassValidator (`go build <target>`)
- [x] 4.3 Implement TestPassValidator (`go test <target>`)
- [x] 4.4 Implement FileChangedValidator (`git diff --name-only` pattern match)
- [x] 4.5 Implement ArtifactExistsValidator (`os.Stat`)
- [x] 4.6 Implement CommandPassValidator (arbitrary command + expected exit code)
- [x] 4.7 Implement OrchestratorApprovalValidator (always returns failed — requires explicit approval)

## 5. Agent Tools

- [x] 5.1 Create `internal/runledger/tools.go` with role-based access control (checkRole)
- [x] 5.2 Implement run_create — planner JSON parsing, validation, run_id generation, journal events
- [x] 5.3 Implement run_read — snapshot retrieval with caching
- [x] 5.4 Implement run_active — active/next executable step query
- [x] 5.5 Implement run_note — scratchpad read/write via journal
- [x] 5.6 Implement run_propose_step_result — result proposal with evidence (no completion)
- [x] 5.7 Implement run_apply_policy — 7 policy actions (retry, decompose, change_agent, change_validator, skip, abort, escalate)
- [x] 5.8 Implement run_approve_step — explicit orchestrator approval for orchestrator_approval steps
- [x] 5.9 Implement run_resume — paused run resumption via ResumeManager

## 6. Resume Protocol

- [x] 6.1 Create `internal/runledger/resume.go` — ResumeManager with staleTTL
- [x] 6.2 Implement FindCandidates with session key filtering and staleness checking
- [x] 6.3 Implement DetectResumeIntent (Korean: 계속, 이어서, 마저; English: resume, continue)
- [x] 6.4 Implement Resume with status validation (ErrRunNotPaused)

## 7. Workspace Isolation

- [x] 7.1 Create `internal/runledger/workspace.go` — WorkspaceManager
- [x] 7.2 Implement NeedsIsolation (build_pass, test_pass, file_changed → true)
- [x] 7.3 Implement CheckDirtyTree, CreateWorktree, RemoveWorktree
- [x] 7.4 Implement ExportPatch (`git format-patch`) and ApplyPatch (`git am`) — no auto-merge

## 8. Rollout & WriteThrough

- [x] 8.1 Create `internal/runledger/writethrough.go` — RolloutStage enum (Shadow, WriteThrough, AuthoritativeRead, ProjectionRetired)
- [x] 8.2 Implement RolloutConfig with stage query methods

## 9. Ent Schemas

- [x] 9.1 Create `internal/ent/schema/run_journal.go` — UUID PK, run_id+seq unique index, enum type, text payload
- [x] 9.2 Create `internal/ent/schema/run_snapshot.go` — unique run_id, status enum, snapshot_data text, last_journal_seq
- [x] 9.3 Create `internal/ent/schema/run_step.go` — run_id+step_id unique index, status enum, evidence/validator as JSON text
- [x] 9.4 Run `go generate ./internal/ent/`

## 10. Configuration & App Integration

- [x] 10.1 Create `internal/config/types_runledger.go` — RunLedgerConfig (enabled, shadow, writeThrough, authoritativeRead, staleTtl, validatorTimeout, plannerMaxRetries, maxRunHistory)
- [x] 10.2 Add RunLedger field to root Config struct in `internal/config/types.go`
- [x] 10.3 Add ProvidesRunLedger to `internal/appinit/module.go`
- [x] 10.4 Create `internal/app/modules_runledger.go` — runLedgerModule with catalog entry "runledger"
- [x] 10.5 Register module in `internal/app/app.go` builder
- [x] 10.6 Add RunLedgerStore and RunLedgerPEV to App struct and populateAppFields

## 11. Testing

- [x] 11.1 types_test.go — marshalPayload, validator type constants, step JSON roundtrip (3 tests)
- [x] 11.2 snapshot_test.go — materialize empty/basic, NextExecutableStep, AllStepsTerminal, ToSummary, ApplyTail, policy retry/decompose/abort (10 tests)
- [x] 11.3 store_test.go — append/retrieve, GetJournalEventsSince, materialize, validation recording, caching, ListRuns, not found (7 tests)
- [x] 11.4 planparse_test.go — fenced/raw/no/invalid JSON, validation (goal, steps, duplicate, agent, validator, cycle, dependency, missing type), ConvertPlanToRunData, inferToolProfile (14 tests)
- [x] 11.5 pev_test.go — verify pass/fail, unknown validator, orchestrator_approval never auto-pass (4 tests)
- [x] 11.6 tools_test.go — BuildTools count, end-to-end create/read/active/note/propose, invalid plan, retry policy, approval, resume intent, planner JSON roundtrip (7 tests)
- [x] 11.7 resume_test.go — FindCandidates, Resume, Resume not paused, intent detection, step summary (6 tests)

## 12. Build & Test Verification

- [x] 12.1 `go build ./...` passes with zero errors
- [x] 12.2 `go test ./internal/runledger/` — all 51 tests pass
- [x] 12.3 `go test ./...` — zero FAIL across entire project
- [x] 12.4 Create `openspec/specs/run-ledger/spec.md` following OpenSpec format
