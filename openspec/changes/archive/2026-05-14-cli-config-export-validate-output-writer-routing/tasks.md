## 1. Output Routing

- [x] 1.1 Route `lango config export` JSON output through the Cobra command writer and warning through `cmd.ErrOrStderr()`.
- [x] 1.2 Route `lango config validate` success output through the Cobra command writer.
- [x] 1.3 Add command-level capture tests for export and validate.

## 2. Spec Sync

- [x] 2.1 Record the export/validate output-writer contract in `config-cli-commands`.
- [x] 2.2 Update downstream `docs/cli/core.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/configcmd -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-config-export-validate-output-writer-routing --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
