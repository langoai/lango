## 1. Help Surface Fix

- [x] 1.1 Update Mission Control proposed-row short help so actionable navigation bindings remain visible alongside `Enter` accept and `d` dismiss.
- [x] 1.2 Add a regression covering a proposed mission in a multi-row missions lane.

## 2. Downstream Sync

- [x] 2.1 Update Mission Control feature docs to make the combined proposed-row help surface explicit.
- [x] 2.2 Update the `cockpit-pages` OpenSpec delta with the preserved-navigation contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate missioncontrol-proposal-help-preserves-navigation --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
