## 1. Command Streams

- [ ] 1.1 Route `lango security secrets list` table and JSON output through the Cobra command writer.
- [ ] 1.2 Route `lango security secrets set` success output through the Cobra command writer.
- [ ] 1.3 Route `lango security secrets delete` prompt/result output through the Cobra command streams.
- [ ] 1.4 Add command-level capture tests using a persistent temp DB-backed bootloader.

## 2. Spec Sync

- [ ] 2.1 Record the secrets command stream contract in `cli-secrets-management`.
- [ ] 2.2 Update downstream `docs/cli/security.md` to match runtime behavior.

## 3. Verification

- [ ] 3.1 Run `go test ./internal/cli/security -count=1`.
- [ ] 3.2 Run `go build ./...`.
- [ ] 3.3 Run `go test ./...`.
- [ ] 3.4 Run `openspec validate cli-security-secrets-stream-routing --strict`.
- [ ] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
