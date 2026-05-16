## 1. Output Routing

- [x] 1.1 Route `lango agent status` table output through the Cobra command writer.
- [x] 1.2 Route `lango agent status --json` output through the Cobra command writer.
- [x] 1.3 Update command-level tests to capture the writer directly.

## 2. Spec Sync

- [x] 2.1 Record the status output-writer contract in `cli-agent-inspection`.
- [x] 2.2 Update downstream CLI docs to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/agent -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-agent-status-output-writer-routing --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
