## 1. Tests First

- [x] 1.1 Add failing sandbox worker tests for decode failure, unregistered tool, tool error, and success exit-code paths.
- [x] 1.2 Add/update a failing `cmd/lango` worker-mode test proving the worker seam exit code is returned.

## 2. Implementation

- [x] 2.1 Add a testable sandbox worker function that accepts injected IO and returns an exit code.
- [x] 2.2 Update `RunWorker` and `cmd/lango` worker-mode wiring to preserve process behavior through the main exit seam.
- [x] 2.3 Preserve existing JSON result protocol and exit-code semantics.

## 3. Verification

- [x] 3.1 Run focused sandbox and `cmd/lango` tests.
- [x] 3.2 Run `openspec validate sandbox-worker-exit-code --strict`.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Sync/archive the OpenSpec change after verification.
