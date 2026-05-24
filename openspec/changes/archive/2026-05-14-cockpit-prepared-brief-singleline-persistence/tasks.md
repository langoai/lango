## 1. Prepared Brief Hardening

- [x] 1.1 Collapse prepared-brief segments into one single-line persisted description.
- [x] 1.2 Add regression coverage for embedded newline removal in accepted descriptions.

## 2. Spec Sync

- [x] 2.1 Record the single-line prepared-brief persistence contract in `mission-control-tui`.
- [x] 2.2 Update downstream cockpit docs to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cockpit-prepared-brief-singleline-persistence --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
