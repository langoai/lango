## 1. Planning And Baseline

- [x] 1.1 Measure the current non-Ent coverage baseline
- [x] 1.2 Identify top files by uncovered statement count
- [x] 1.3 Validate and commit this umbrella OpenSpec plan

## 2. Coverage Measurement Contract

- [ ] 2.1 Add a repeatable coverage report command or script for non-generated Go code
- [ ] 2.2 Exclude generated code by known generated paths and generated-file markers
- [ ] 2.3 Add tests for the coverage filtering logic
- [ ] 2.4 Report total coverage, covered statements, total statements, uncovered statements, and top uncovered files
- [ ] 2.5 Add an executable threshold gate that fails when non-generated coverage is below 90%

## 3. Coverage Wave: Small Deterministic Packages

- [ ] 3.1 Add focused tests for low-risk pure/helper packages with low coverage
- [ ] 3.2 Add meaningful assertions for enum, parser, embed, and simple wrapper behavior
- [ ] 3.3 Commit the wave separately

## 4. Coverage Wave: CLI/TUI State And Command Surfaces

- [ ] 4.1 Add targeted tests for `internal/cli/tuicore/state_update.go`
- [ ] 4.2 Add targeted tests for `internal/cli/settings/editor.go` and `internal/cli/settings/menu.go`
- [ ] 4.3 Cover command construction, output seams, error paths, and state transitions
- [ ] 4.4 Commit the wave separately

## 5. Coverage Wave: Storage And Knowledge

- [ ] 5.1 Add focused tests for `internal/storage/facade.go`
- [ ] 5.2 Add focused tests for `internal/storagebroker/server.go` and `internal/storagebroker/client.go`
- [ ] 5.3 Add focused tests for `internal/knowledge/store.go`
- [ ] 5.4 Cover CRUD, error handling, unavailable backends, and boundary conditions
- [ ] 5.5 Commit the wave separately

## 6. Coverage Wave: Runtime And Workflow

- [ ] 6.1 Add deterministic tests or seams for `internal/adk/agent.go`
- [ ] 6.2 Add deterministic tests for `internal/workflow/engine.go`
- [ ] 6.3 Add deterministic tests for `internal/turnrunner/runner.go`
- [ ] 6.4 Add persistence boundary tests for `internal/turntrace/store.go`
- [ ] 6.5 Cover workflow execution/status paths, turn-runner outcomes, and trace persistence boundaries
- [ ] 6.6 Commit the wave separately

## 7. Coverage Wave: App Wiring And P2P

- [ ] 7.1 Add focused tests for `internal/app/wiring_p2p.go`
- [ ] 7.2 Add focused tests for `internal/app/wiring.go`
- [ ] 7.3 Add focused tests for `internal/app/tools_meta.go`
- [ ] 7.4 Add focused tests for `internal/p2p/handshake/handshake.go`
- [ ] 7.5 Commit the wave separately

## 8. Verification And Enforcement

- [ ] 8.1 Run `go build ./...`
- [ ] 8.2 Run `go test ./...`
- [ ] 8.3 Run the non-generated coverage report and confirm coverage is at least 90%
- [ ] 8.4 Run the executable 90% coverage gate and confirm it passes
- [ ] 8.5 Run `git diff --check`
- [ ] 8.6 Run subagent-driven spec and code-quality review checkpoints
- [ ] 8.7 Run `openspec validate --all --strict`
- [ ] 8.8 Archive the completed OpenSpec change
