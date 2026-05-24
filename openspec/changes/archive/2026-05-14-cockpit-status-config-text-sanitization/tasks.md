## 1. Status Page Config Hardening

- [x] 1.1 Sanitize rendered provider and model values in the System section.
- [x] 1.2 Add regression coverage for malformed config-fed system labels.

## 2. Spec Sync

- [x] 2.1 Record the expanded Status page text-sanitization contract in `cockpit-status-page`.
- [x] 2.2 Update downstream cockpit docs to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cockpit-status-config-text-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
