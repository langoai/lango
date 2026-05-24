## Why

The broker boundary has improved, but production code can still regress if raw storage escape hatches or direct Ent extraction are reintroduced. This change adds explicit enforcement so the current boundary hardening stays mechanically protected.

## What Changes

- Add architecture tests that fail when removed raw storage facade accessors reappear in production packages.
- Add architecture tests that fail when payment production wiring reintroduces direct `session.EntStore.Client()` extraction.
- Keep broker transport regression tests as an explicit release gate for large payload handling, write serialization, and graceful shutdown.
- Remove the remaining production fallback that reconstructs payment transaction persistence from `session.EntStore.Client()`.

## Capabilities

### New Capabilities

- `broker-transport-regression-gate`: Explicit regression gate covering concurrent write serialization, large response decode, and graceful shutdown semantics.

### Modified Capabilities

- `architecture-boundary-enforcement`: Production package boundary enforcement expands to broker/storage raw-handle regressions.
- `brokered-storage`: Broker-backed storage boundary now includes explicit enforcement against removed raw accessors and payment-side Ent extraction regressions.

## Impact

- Affected code:
  - `internal/archtest/*`
  - `internal/app/wiring_payment.go`
  - `internal/storagebroker/*`
  - `docs/architecture/*`
- Systems:
  - Production app/CLI boundary enforcement
  - Broker transport regression coverage
