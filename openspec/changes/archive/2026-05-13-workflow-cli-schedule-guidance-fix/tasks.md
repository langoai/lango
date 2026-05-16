## 1. CLI Guidance Fix

- [x] 1.1 Replace nonexistent workflow registration endpoint guidance in `workflow run --schedule`.
- [x] 1.2 Add a CLI regression covering the new scheduled-run message.

## 2. Docs And Spec Sync

- [x] 2.1 Update CLI automation docs for the new scheduled-run guidance.
- [x] 2.2 Update workflow management spec to the same contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/workflow -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
