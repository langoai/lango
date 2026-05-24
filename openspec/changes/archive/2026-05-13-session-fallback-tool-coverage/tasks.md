## 1. Session Fallback Coverage

- [x] 1.1 Add tool-entrypoint regression coverage for observational-memory session fallback.
- [x] 1.2 Add tool-entrypoint regression coverage for librarian pending-inquiry session fallback.

## 2. Spec Sync

- [x] 2.1 Update observational-memory and proactive-librarian specs for the current-session fallback contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/memory ./internal/librarian -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
