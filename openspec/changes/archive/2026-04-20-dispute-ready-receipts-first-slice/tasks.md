## 1. Receipt Domain And Storage Model

- [x] 1.1 Add the `internal/receipts` package with submission receipt, transaction receipt, event types, and lite provenance summary
- [x] 1.2 Add receipt store helpers for submission creation, transaction linkage, current-submission pointer, and event trail append

## 2. Minimal Receipt Integration Surface

- [x] 2.1 Add a narrow `create_dispute_ready_receipt` meta tool for artifact submissions

## 3. Operator Surface And Docs

- [x] 3.1 Add dispute-ready receipt lite docs and truth-align surrounding surfaces

## 4. Verification And OpenSpec Closeout

- [x] 4.1 Run targeted tests while implementing each slice
- [x] 4.2 Run `go test ./...`, `go build ./...`, and `python3 -m mkdocs build --strict`
- [x] 4.3 Sync main specs and archive the completed change
