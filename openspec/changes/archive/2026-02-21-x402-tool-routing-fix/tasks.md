## 1. Code Changes

- [x] 1.1 Rename tool name from `x402_fetch` to `payment_x402_fetch` in `internal/tools/payment/payment.go:257`
- [x] 1.2 Update approval description switch case from `x402_fetch` to `payment_x402_fetch` in `internal/app/tools.go:1796`

## 2. Documentation Updates

- [x] 2.1 Update tool table and protocol flow in `README.md` (lines 542, 556)
- [x] 2.2 Update X402 protocol flow in `docs/payments/x402.md` (line 18)
- [x] 2.3 Update USDC tool table in `docs/payments/usdc.md` (line 23)
- [x] 2.4 Update X402 V2 spec in `openspec/specs/x402-v2/spec.md` (lines 7, 34)

## 3. Verification

- [x] 3.1 Run `go build ./...` — verify no compilation errors
- [x] 3.2 Run `go test ./...` — verify all tests pass
- [x] 3.3 Verify `payment_x402_fetch` routes to vault agent via `payment_` prefix match
