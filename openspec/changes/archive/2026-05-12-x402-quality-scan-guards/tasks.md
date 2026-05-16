## 1. x402 Static Quality Gates

- [x] 1.1 Add a package-source test that rejects `context.TODO()` inside `internal/x402`.
- [x] 1.2 Add a codebase scan test that rejects legacy `NewX402Client` references.

## 2. Verification

- [x] 2.1 Run `go test ./internal/x402 -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.
- [ ] 2.4 Validate and archive the OpenSpec change.
