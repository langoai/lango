## Why

The payment service already rejects malformed USDC amounts, but the current regression only checks for a high-level `invalid amount` substring. That leaves the parsing cause free to disappear, even though the operator-facing error is more useful when it preserves the exact reason.

## What Changes

- Tighten `payment.Service.Send` tests so invalid amount errors must preserve the wrapped parsing cause and rejected input.
- Cover both malformed input and too-many-decimals paths.
- Sync the production-readiness spec to make that error shape explicit.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: invalid amount failures now have a direct regression for wrapped parsing-cause preservation.

## Impact

- Affected code: `internal/payment/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
