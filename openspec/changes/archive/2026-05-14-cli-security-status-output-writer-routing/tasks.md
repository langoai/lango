## 1. Output Routing

- [x] 1.1 Route `lango security status` table output through the Cobra command writer.
- [x] 1.2 Route `lango security status --json` output through the Cobra command writer.
- [x] 1.3 Update status rendering tests to capture buffers directly.

## 2. Spec Sync

- [x] 2.1 Record the security status output-writer contract in `cli-security-status`.
- [x] 2.2 Update downstream `docs/cli/security.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/security -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-security-status-output-writer-routing --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
