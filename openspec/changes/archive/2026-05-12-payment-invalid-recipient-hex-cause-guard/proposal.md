## Why

The invalid recipient regression now covers malformed address formats, but it still does not prove that an address with the right length and prefix but invalid hex characters preserves the specific hex-validation cause. That distinction matters when debugging bad upstream payment routing.

## What Changes

- Extend `payment.Service.Send` invalid recipient coverage to include the invalid-hex path.
- Sync the production-readiness spec so invalid recipient failures explicitly cover both format and hex validation causes.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: invalid recipient regressions now cover both malformed-format and invalid-hex causes.

## Impact

- Affected code: `internal/payment/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
