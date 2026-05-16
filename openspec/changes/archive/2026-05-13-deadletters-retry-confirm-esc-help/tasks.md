## 1. UX Fix

- [x] 1.1 Expose `Esc` in Dead Letters help while retry confirmation is pending.
- [x] 1.2 Add regression coverage for the confirm-state help contract.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit-pages spec to require confirm-cancel help exposure.
- [x] 2.2 Update cockpit feature docs to describe `Esc` as the retry-confirm cancel key.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
