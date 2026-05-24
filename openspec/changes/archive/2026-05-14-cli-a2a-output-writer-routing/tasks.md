## 1. Output Routing

- [x] 1.1 Route `lango a2a card` table and JSON output through the Cobra command writer.
- [x] 1.2 Route `lango a2a check` table and JSON output through the Cobra command writer.
- [x] 1.3 Add command-level capture tests for local and remote card inspection.

## 2. Spec Sync

- [x] 2.1 Record the A2A CLI output-writer contract in `cli-a2a-management`.
- [x] 2.2 Update downstream `docs/cli/a2a.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/a2a -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-a2a-output-writer-routing --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
