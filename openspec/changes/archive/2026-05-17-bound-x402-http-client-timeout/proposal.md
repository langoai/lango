## Why

The X402 interceptor wraps a plain `http.Client` with no timeout. `payment_x402_fetch` passes request contexts, but the underlying client should also have a bounded transport deadline so paid HTTP requests cannot hang indefinitely if callers provide a long-lived context or an upstream stalls.

## What Changes

- Give the X402 wrapped base HTTP client a finite default timeout.
- Preserve lazy initialization and cached client behavior.
- Add regression coverage proving the returned X402 HTTP client is bounded.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `x402-v2`: X402 payment HTTP requests use a bounded default HTTP client timeout.

## Impact

- Affected code: `internal/x402/interceptor.go`.
- Affected tests: `internal/x402`.
- Affected specs: `openspec/specs/x402-v2/spec.md`.
