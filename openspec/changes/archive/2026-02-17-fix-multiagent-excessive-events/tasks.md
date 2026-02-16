## 1. A2A Wiring Order Fix

- [x] 1.1 Move A2A remote agent loading before `BuildAgentTree()` in `internal/app/wiring.go`

## 2. Author Identity Fix

- [x] 2.1 Add `Author string` field to `session.Message` in `internal/session/store.go`
- [x] 2.2 Add `author` field to ent Message schema in `internal/ent/schema/message.go`
- [x] 2.3 Run `go generate ./internal/ent` to regenerate ent code
- [x] 2.4 Add `rootAgentName` to `SessionAdapter` and `EventsAdapter` in `internal/adk/state.go`
- [x] 2.5 Update `EventsAdapter.All()` to use stored author with rootAgentName fallback
- [x] 2.6 Add `rootAgentName` to `SessionServiceAdapter` in `internal/adk/session_service.go`
- [x] 2.7 Save `evt.Author` in `AppendEvent()` to persist author in messages
- [x] 2.8 Pass agent name to `NewSessionServiceAdapter` in `internal/adk/agent.go`
- [x] 2.9 Update `ent_store.go` to save and load the `author` field

## 3. Orchestrator Prompt Optimization

- [x] 3.1 Add `MaxDelegationRounds` field to `orchestration.Config`
- [x] 3.2 Update orchestrator instruction with short-circuit directive for simple queries
- [x] 3.3 Add max delegation rounds mention in orchestrator prompt
- [x] 3.4 Set default `MaxDelegationRounds: 3` in `internal/app/wiring.go`

## 4. Conditional Sub-Agent Creation

- [x] 4.1 Refactor `BuildAgentTree` to only create sub-agents with assigned tools (planner always)
- [x] 4.2 Generate orchestrator instruction dynamically from actual sub-agent list

## 5. Test Updates

- [x] 5.1 Update `state_test.go` with rootAgentName parameter and multi-agent author test
- [x] 5.2 Update `session_service_test.go` with rootAgentName parameter and author preservation test
- [x] 5.3 Update `orchestrator_test.go` with conditional sub-agent tests (no memory, no tools)

## 6. Verification

- [x] 6.1 Run `go build ./...` — passes
- [x] 6.2 Run `go test ./internal/adk/... ./internal/orchestration/... ./internal/session/... ./internal/app/...` — all pass
