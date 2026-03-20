## 1. Runtime Workspace Activation

- [x] 1.1 Enable workspace manager wiring behind the phase/stage gate
- [ ] 1.2 Add app/runtime tests covering isolated validation paths

## 2. Retry-Safe Isolation

- [ ] 2.1 Guarantee unique branch/worktree identity per isolated attempt
- [ ] 2.2 Ensure cleanup removes both worktree and branch metadata
- [ ] 2.3 Add repeated-validation tests

## 3. Tool Governance

- [ ] 3.1 Narrow execution tools to the step's `ToolProfile`
- [ ] 3.2 Keep supervisor/orchestrator profiles minimal
- [ ] 3.3 Add tests for tool visibility by profile

## 4. Verification

- [ ] 4.1 Run `go build ./...`
- [ ] 4.2 Run RunLedger/orchestration/tool tests
- [ ] 4.3 Run `go test ./...`
