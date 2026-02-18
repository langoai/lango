# Tasks

## Implementation

- [x] Step 1: Add Coinbase X402 Go SDK dependency (`go get github.com/coinbase/x402/go`)
- [x] Step 2: Create `internal/x402/signer.go` — SignerProvider bridge (SecretsStore → SDK ClientEvmSigner)
- [x] Step 3: Rewrite `internal/x402/types.go` — Remove V1 types, add CAIP2Network + Config
- [x] Step 4: Rewrite `internal/x402/handler.go` — SDK-based X402Client factory
- [x] Step 5: Rewrite `internal/x402/interceptor.go` — V2 SDK-based interceptor with spending limit hooks
- [x] Step 6: Update `internal/payment/service.go` — Remove HandleX402, add RecordX402Payment
- [x] Step 7: Update `internal/payment/types.go` — Remove X402Challenge, add X402PaymentRecord
- [x] Step 8: Add `x402_fetch` tool to `internal/tools/payment/payment.go`
- [x] Step 9: Wire X402 into `internal/app/` (wiring.go, app.go, tools.go, types.go)
- [x] Step 10: Add `payment_method` enum to PaymentTx Ent schema + regenerate

## Verification

- [x] `go build ./...` — Compile check
- [x] `go test ./internal/x402/...` — X402 package tests
- [x] `go test ./internal/payment/...` — Payment package tests
- [x] `go test ./...` — Full test suite
