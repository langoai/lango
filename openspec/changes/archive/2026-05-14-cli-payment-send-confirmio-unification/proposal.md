## Why

`lango payment send` still carries its own confirmation parser even though the repository now has a shared confirmation helper with stream seams. That duplicate prompt path makes a high-stakes payment flow less consistent than the surrounding security and config commands.

## What Changes

- Replace the inline `payment send` confirmation parser with the shared `prompt.ConfirmIO(...)` helper
- Preserve the existing non-interactive `--force` safeguard and the multi-line payment summary shown before confirmation
- Add payment CLI regressions for abort, force, and non-interactive guidance
- Clarify the command-stream confirmation contract in specs/docs

## Capabilities

### New Capabilities

### Modified Capabilities
- `payment-service`: `lango payment send` confirmation uses the shared helper and Cobra command streams

## Impact

- Affected code: `internal/cli/payment/send.go`, `internal/cli/payment/payment_test.go`
- Affected docs/specs: `docs/payments/usdc.md`, `openspec/specs/payment-service/spec.md`
- No feature expansion; this is a consistency and testability improvement
