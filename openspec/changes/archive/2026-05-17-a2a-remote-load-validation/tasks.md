## 1. Tests First

- [x] 1.1 Add failing `internal/a2a` tests for missing `agentCardUrl` and mixed valid/invalid remote loading.

## 2. Implementation

- [x] 2.1 Update `LoadRemoteAgents` to return partial successful agents plus an aggregate error for skipped configured remotes.
- [x] 2.2 Preserve app startup graceful degradation and existing caller warning behavior.

## 3. Downstream Artifacts

- [x] 3.1 Update A2A public docs to distinguish partial remote loading warnings from successful remote wiring.

## 4. Verification

- [x] 4.1 Run focused A2A tests.
- [x] 4.2 Run `openspec validate a2a-remote-load-validation --strict`.
- [x] 4.3 Run `go build ./...` and `go test ./...`.
- [x] 4.4 Sync/archive the OpenSpec change after verification.
