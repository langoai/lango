## 1. Footer Hint Sync

- [x] 1.1 Update Mission Control footer hints so a focused proposed mission advertises accept/dismiss actions.
- [x] 1.2 Add a regression covering the proposal-focused footer hint.

## 2. Downstream Sync

- [x] 2.1 Update public cockpit docs so Mission Control guidance mentions the proposal-focused footer hint.
- [x] 2.2 Update the `cockpit-pages` OpenSpec delta with the footer-hint contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate missioncontrol-proposal-footer-hint-sync --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
