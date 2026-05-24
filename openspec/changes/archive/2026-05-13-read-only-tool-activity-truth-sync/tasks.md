## 1. Capability Fix

- [x] 1.1 Reclassify cron inspection tools as query activity.
- [x] 1.2 Add explicit read/query activity metadata to workspace and git inspection tools.
- [x] 1.3 Add regressions that lock the updated capability metadata.

## 2. Docs And Spec Sync

- [x] 2.1 Update prompt/docs wording to describe the inspection paths consistently.
- [x] 2.2 Add OpenSpec coverage for the read/query capability classification.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cron ./internal/app -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
