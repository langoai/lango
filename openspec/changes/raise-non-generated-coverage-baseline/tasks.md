## 1. Planning And Baseline

- [x] 1.1 Measure the current non-Ent coverage baseline
- [x] 1.2 Identify top files by uncovered statement count
- [x] 1.3 Validate and commit this umbrella OpenSpec plan

## 2. Coverage Measurement Contract

- [x] 2.1 Add a repeatable coverage report command or script for non-generated Go code
- [x] 2.2 Exclude generated code by known generated paths and generated-file markers
- [x] 2.3 Add tests for the coverage filtering logic
- [x] 2.4 Report total coverage, covered statements, total statements, uncovered statements, and top uncovered files
- [x] 2.5 Add an executable threshold gate that fails when non-generated coverage is below 90%

## 3. Coverage Wave: Small Deterministic Packages

- [x] 3.1 Add focused tests for low-risk pure/helper packages with low coverage
- [x] 3.2 Add meaningful assertions for enum, parser, embed, and simple wrapper behavior
- [x] 3.3 Commit the wave separately

## 4. Coverage Wave: CLI/TUI State And Command Surfaces

- [x] 4.1 Add targeted tests for `internal/cli/tuicore/state_update.go`
- [x] 4.2 Add targeted tests for `internal/cli/settings/editor.go` and `internal/cli/settings/menu.go`
- [x] 4.3 Cover command construction, output seams, error paths, and state transitions
- [x] 4.4 Commit the wave separately

## 5. Coverage Wave: Storage And Knowledge

- [x] 5.1 Add focused tests for `internal/storage/facade.go`
- [x] 5.2 Add focused tests for `internal/storagebroker/server.go` and `internal/storagebroker/client.go`
- [x] 5.3 Add focused tests for `internal/knowledge/store.go`
- [x] 5.4 Cover CRUD, error handling, unavailable backends, and boundary conditions
- [x] 5.5 Commit the wave separately

## 6. Coverage Wave: Runtime And Workflow

- [x] 6.1 Add deterministic tests or seams for `internal/adk/agent.go`
- [x] 6.2 Add deterministic tests for `internal/workflow/engine.go`
- [x] 6.3 Add deterministic tests for `internal/turnrunner/runner.go`
- [x] 6.4 Add persistence boundary tests for `internal/turntrace/store.go`
- [x] 6.5 Cover workflow execution/status paths, turn-runner outcomes, and trace persistence boundaries
- [x] 6.6 Commit the wave separately

## 7. Coverage Wave: App Wiring And P2P

- [x] 7.1 Add focused tests for `internal/app/wiring_p2p.go`
- [x] 7.2 Add focused tests for `internal/app/wiring.go`
- [x] 7.3 Add focused tests for `internal/app/tools_meta.go`
- [x] 7.4 Add focused tests for `internal/p2p/handshake/handshake.go`
- [x] 7.5 Commit the wave separately

## 8. Coverage Wave: High-Impact Service Boundaries

- [x] 8.1 Add focused tests for remaining `internal/storagebroker/server.go` dispatch, payload protection, and error paths
- [x] 8.2 Add focused tests for remaining `internal/skill/importer.go` import validation, filesystem, and failure paths
- [x] 8.3 Add focused tests for remaining `internal/receipts/store.go` receipt lifecycle and query boundaries
- [x] 8.4 Commit the wave separately

## 9. Coverage Wave: App And CLI Execution Surfaces

- [x] 9.1 Add focused tests for remaining `internal/app/app.go`, `internal/app/modules.go`, and app wiring branches
- [x] 9.2 Add focused tests for `cmd/lango/main.go` command routing and error exits through seams
- [x] 9.3 Add focused tests for `internal/cli/chat/chat.go`, `internal/cli/settings/editor.go`, `internal/cli/sandbox/sandbox.go`, and `internal/cli/onboard/wizard.go`
- [x] 9.4 Commit the wave separately

## 10. Coverage Wave: Remaining P2P And Payment Surfaces

- [x] 10.1 Add focused tests for remaining `internal/p2p/handshake/handshake.go` stream and approval branches
- [x] 10.2 Add focused tests for high-uncovered P2P protocol, settlement, discovery, and payment paths that do not require external services
- [x] 10.3 Add focused tests for remaining `internal/app/tools_escrow.go` and payment tool wrappers
- [x] 10.4 Commit the wave separately

## 11. Coverage Wave: Cron, Broker, And App Helpers

- [x] 11.1 Add focused tests for `internal/cron/store.go` Ent-backed store and conversion helpers
- [x] 11.2 Add focused tests for remaining `internal/storagebroker/server.go` and `internal/storagebroker/client.go` RPC wrapper branches
- [x] 11.3 Add focused tests for app P2P helper surfaces that do not require external services
- [x] 11.4 Commit the wave separately

## 12. Coverage Wave: Smart Account, Wallet, And Git Bundle

- [x] 12.1 Add focused tests for `internal/wallet/spending.go` limiter behavior and store error paths
- [x] 12.2 Add focused tests for `internal/smartaccount/manager.go` manager helpers and encoding branches
- [x] 12.3 Add focused tests for `internal/p2p/gitbundle/protocol.go` protocol handler dispatch and error branches
- [x] 12.4 Add focused tests for app Smart Account tools and wiring helper branches
- [x] 12.5 Commit the wave separately

## 13. Coverage Wave: Filesystem, Test Utilities, And Knowledge Branches

- [x] 13.1 Add focused tests for `internal/tools/filesystem/filesystem.go` delete, exists, mkdir, copy, and path access branches
- [x] 13.2 Add focused tests for `internal/testutil/mock_session_store.go` CRUD, salt, end, counters, copy semantics, and error injection
- [x] 13.3 Add focused tests for `internal/knowledge/store.go` FTS5 result resolution, learning lookup branches, and low-risk search paths
- [x] 13.4 Commit the wave separately

## 14. Coverage Wave: App P2P, Meta Tools, And Knowledge Store

- [x] 14.1 Add focused tests for deterministic `internal/app/wiring_p2p.go` helper and wiring branches
- [x] 14.2 Add focused tests for deterministic `internal/app/tools_meta.go` tool handlers and helper branches
- [x] 14.3 Add focused tests for remaining low-risk `internal/knowledge/store.go` scoring, stats, external ref, and deletion branches
- [x] 14.4 Run subagent-driven review checkpoints for the wave
- [x] 14.5 Commit the wave separately

## 15. Coverage Wave: CLI Entrypoints, Chat State, Settings Navigation, And Config Redaction

- [x] 15.1 Add focused tests for deterministic `cmd/lango/main.go` helper and adapter branches
- [x] 15.2 Add focused tests for deterministic `internal/cli/chat/chat.go` state, accessor, session-mode, and key-helper branches
- [x] 15.3 Add focused tests for deterministic `internal/cli/configcmd/getset.go` redaction, scalar, reflection, formatting, and set-field branches
- [x] 15.4 Add focused tests for deterministic `internal/cli/settings/editor.go` dependency navigation and setup-flow branches
- [x] 15.5 Run subagent-driven review checkpoints for the wave
- [ ] 15.6 Commit the wave separately

## 16. Verification And Enforcement

- [ ] 16.1 Run `go build ./...`
- [ ] 16.2 Run `go test ./...`
- [ ] 16.3 Run the non-generated coverage report and confirm coverage is at least 90%
- [ ] 16.4 Run the executable 90% coverage gate and confirm it passes
- [ ] 16.5 Run `git diff --check`
- [ ] 16.6 Run subagent-driven spec and code-quality review checkpoints
- [ ] 16.7 Run `openspec validate --all --strict`
- [ ] 16.8 Archive the completed OpenSpec change
