## 1. Tests

- [x] 1.1 Add failing root CLI coverage for `lango bg` boundary error copy.
- [x] 1.2 Add failing bg command help coverage for in-process manager scope.
- [x] 1.3 Add failing public docs guard for README, CLI index, and background automation docs caveats.

## 2. Implementation

- [x] 2.1 Update root `lango bg` provider error to describe the in-memory server boundary accurately.
- [x] 2.2 Update `internal/cli/bg` long/help copy without breaking in-process manager-backed commands.
- [x] 2.3 Update public docs where `lango bg` commands are listed.

## 3. Verification

- [x] 3.1 Run focused bg/root/docs tests.
- [x] 3.2 Run full `go build ./...`, `go test ./...`, `git diff --check`, and strict OpenSpec validation.
- [x] 3.3 Run subagent-driven review and address required findings.
- [x] 3.4 Sync specs, archive the change, and commit the scoped unit.
