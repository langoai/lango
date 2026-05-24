## 1. Regression Coverage

- [x] 1.1 Add exact missing-parameter regressions for `agent_spawn`, `agent_wait`, and `agent_stop`.
- [x] 1.2 Add exact missing-parameter regressions for `task_create`, `task_get`, and `task_update`.

## 2. Docs And Spec Sync

- [x] 2.1 Update agent-facing prompt/docs for the required input contract.
- [x] 2.2 Sync the control-plane and production-readiness specs to the same contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/agentrt -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
