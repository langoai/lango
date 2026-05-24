## 1. Output Routing

- [x] 1.1 Route `lango run list` output through the Cobra command writer.
- [x] 1.2 Route `lango run status` and `lango run journal` output through the Cobra command writer.
- [x] 1.3 Add command-level capture tests for list, status, and journal.

## 2. Spec Sync

- [x] 2.1 Record the RunLedger CLI output-writer contract in `run-ledger`.
- [x] 2.2 Update downstream `docs/cli/run.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/run -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-run-output-writer-routing --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
