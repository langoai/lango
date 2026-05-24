## 1. Output Routing

- [x] 1.1 Route root `lango status` table output through the Cobra command writer.
- [x] 1.2 Add regression coverage that captures root status output via `executeCommand`.

## 2. Spec Sync

- [x] 2.1 Record the root status output-writer contract in `cli-status-dashboard`.
- [x] 2.2 Update downstream `docs/cli/status.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/status -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-status-output-writer-routing --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
