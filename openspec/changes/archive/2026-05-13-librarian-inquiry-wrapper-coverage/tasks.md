## 1. Guard Coverage

- [x] 1.1 Add tool-entrypoint regression coverage for missing `inquiry_id` on `librarian_dismiss_inquiry`.

## 2. Prompt And Docs Sync

- [x] 2.1 Update README and multi-agent docs for the librarian inquiry input contract.
- [x] 2.2 Update proactive-librarian, downstream-docs, and production-readiness specs for the same contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/librarian -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
