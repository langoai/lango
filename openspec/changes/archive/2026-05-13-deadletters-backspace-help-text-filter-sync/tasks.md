## 1. Help Truth Sync

- [x] 1.1 Add a Dead Letters regression for the `Backspace` help label.
- [x] 1.2 Relabel `Backspace` help to describe active text-filter editing.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit-pages spec to require the active text-filter `Backspace` wording.
- [x] 2.2 Update cockpit feature docs to describe the same `Backspace` contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
