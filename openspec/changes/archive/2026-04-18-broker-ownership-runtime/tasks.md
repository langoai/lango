## 1. Reader RPCs

- [x] 1.1 Add broker reader RPCs for learning history, pending inquiries, workflow runs, alerts, and reputation details.
- [x] 1.2 Add broker-backed storage adapters/readers for those runtime surfaces.

## 2. Runtime Reader Adoption

- [x] 2.1 Switch production CLI and app reader paths to broker-backed storage readers.
- [x] 2.2 Remove reader-side production dependence on `FTSDB()` / `PaymentClient()` where this slice can replace them.

## 3. Verification

- [x] 3.1 Add/update tests for broker-backed runtime reader paths.
- [x] 3.2 Run `go build ./...`, `go test ./...`, and `openspec validate --type change broker-ownership-runtime`.
