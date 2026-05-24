## 1. Output Routing

- [x] 1.1 Route `lango config list`, `create`, `use`, `delete`, and `import` output through the Cobra command writer.
- [x] 1.2 Route delete confirmation input/output through `cmd.InOrStdin()` and `cmd.OutOrStdout()`.
- [x] 1.3 Add command-level capture tests for profile-management flows.

## 2. Spec Sync

- [x] 2.1 Record the profile-management output-writer contract in `config-cli-commands`.
- [x] 2.2 Update downstream `docs/cli/core.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/configcmd -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-config-profile-output-writer-routing --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
