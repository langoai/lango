## 1. Dead Letters Hardening

- [x] 1.1 Sanitize rendered backlog-row and degraded-state text.
- [x] 1.2 Sanitize rendered detail-pane and summary-strip text.
- [x] 1.3 Sanitize retry success/failure/follow-up status messages.
- [x] 1.4 Add regression coverage for malformed dead-letter metadata.

## 2. Spec Sync

- [x] 2.1 Record the Dead Letters page text-sanitization contract in a main spec.
- [x] 2.2 Update downstream cockpit docs to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cockpit-deadletters-text-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
