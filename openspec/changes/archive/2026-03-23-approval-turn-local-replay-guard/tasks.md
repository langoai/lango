## 1. Turn-Local Approval State

- [x] 1.1 Add request-scoped approval state helpers in `internal/approval/`
- [x] 1.2 Add canonical `tool + params` keying and replay lookup helpers
- [x] 1.3 Thread turn-local approval state through channel and gateway request contexts

## 2. Approval Outcome Structuring

- [x] 2.1 Add approval sentinel errors and provider-aware wrapping
- [x] 2.2 Update approval middleware to support turn-local grant reuse and replay-blocked negative outcomes
- [x] 2.3 Add structured approval logs across middleware and Telegram provider

## 3. Verification And Prompt Guard

- [x] 3.1 Update browser/navigator prompt guidance to avoid immediate retry after approval failure
- [x] 3.2 Add unit and integration tests for turn-local replay behavior and approval error classification
- [x] 3.3 Run `go build ./...` and `go test ./...`
