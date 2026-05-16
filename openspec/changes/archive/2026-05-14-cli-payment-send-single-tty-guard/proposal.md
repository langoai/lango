## Why

`lango payment send` still carries a second process-global interactive check after the shared TTY-input guard already rejects non-interactive input. That duplication adds unnecessary state and test setup without changing the user-visible contract.

## What Changes

- Remove the redundant `paymentInteractiveCheck` path from `payment send`
- Keep the existing `--force` guidance and confirmation behavior unchanged through the shared TTY-input guard
- Simplify payment CLI tests so they no longer need to override the redundant check

## Capabilities

### New Capabilities

### Modified Capabilities
- `payment-service`: payment send relies on the shared TTY-input guard as its single non-interactive confirmation gate

## Impact

- Affected code: `internal/cli/payment/send.go`, `internal/cli/payment/payment_test.go`
- Affected specs: `openspec/specs/payment-service/spec.md`
- No user-facing behavior change; this is a runtime simplification and test cleanup
