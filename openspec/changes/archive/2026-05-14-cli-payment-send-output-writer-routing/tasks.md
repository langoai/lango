## 1. Output Routing

- [x] 1.1 Route `lango payment send` prompt and success output through the Cobra command writer.
- [x] 1.2 Route confirmation input through `cmd.InOrStdin()` and JSON output through the same writer.
- [x] 1.3 Add command-level capture tests using lightweight send seams.

## 2. Spec Sync

- [x] 2.1 Record the payment send output/input-writer contract in `cli-payment-management`.
- [x] 2.2 Update downstream `docs/cli/payment.md` to match runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/payment -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate cli-payment-send-output-writer-routing --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
