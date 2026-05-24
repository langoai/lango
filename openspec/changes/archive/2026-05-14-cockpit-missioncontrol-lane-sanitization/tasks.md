## 1. Mission Control Lane Hardening

- [x] 1.1 Sanitize rendered mission lane text.
- [x] 1.2 Sanitize rendered decision/activity/agenda lane text.
- [x] 1.3 Add regression coverage for malformed Mission Control lane metadata.

## 2. Spec Sync

- [x] 2.1 Record the Mission Control lane text-sanitization contract in `mission-control-tui`.
- [x] 2.2 Update downstream cockpit docs to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cockpit-missioncontrol-lane-sanitization --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
