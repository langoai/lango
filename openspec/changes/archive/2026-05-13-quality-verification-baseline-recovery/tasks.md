## 1. Quality Baseline Recovery

- [x] 1.1 Confirm the landed proposal/service, Mission Control projector, and parallel executor quality hardening matches the intended baseline.
- [x] 1.2 Record the missing quality baseline OpenSpec change artifacts.

## 2. Archive Hygiene

- [x] 2.1 Mark `2026-05-12-chat-markdown-panic-fallback` archived tasks as complete.
- [x] 2.2 Mark `2026-05-12-missioncontrol-workbench-helper-split` archived tasks as complete.

## 3. Verification

- [x] 3.1 Run `go test ./internal/proposal ./internal/cli/cockpit ./internal/streamx -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate quality-verification-baseline-recovery --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
