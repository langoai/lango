## Why

`payment_x402_fetch` already declares `url` as a required parameter, but its handler still used a custom `"url is required"` branch instead of the standardized wrapper-level missing-parameter contract. That makes one payment tool behave differently from the rest of the hardened tool surface.

## What Changes

- Tighten `payment_x402_fetch` to use `toolparam.RequireString` for `url`.
- Add regression coverage for the missing-URL wrapper path.
- Sync payment specs for the wrapper-level `url` requirement.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `payment-tools`: `payment_x402_fetch` now preserves actionable wrapper-level missing-parameter errors for `url`.
- `production-readiness`: wrapper-level request-guard coverage now includes `payment_x402_fetch`.

## Impact

- Affected code: `internal/tools/payment/payment.go`
- Affected tests: `internal/tools/payment/tools_test.go`
- Affected specs: `openspec/specs/payment-tools/spec.md`, `openspec/specs/production-readiness/spec.md`
