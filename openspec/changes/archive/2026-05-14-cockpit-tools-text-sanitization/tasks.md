## 1. Tools Page Hardening

- [x] 1.1 Sanitize rendered category names and descriptions.
- [x] 1.2 Sanitize rendered tool names, descriptions, and safety labels.
- [x] 1.3 Key safety styling off the sanitized safety label.
- [x] 1.4 Add regression coverage for malformed catalog metadata.

## 2. Spec Sync

- [x] 2.1 Record the Tools page text-sanitization contract in `cockpit-tools-page`.
- [x] 2.2 Update downstream cockpit docs to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cockpit-tools-text-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
