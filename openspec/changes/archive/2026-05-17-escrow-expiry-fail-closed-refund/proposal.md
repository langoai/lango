## Why

Escrow expiry is currently inconsistent across entry points. Explicit `Expire` refunds locked funds for funded or active escrows, but the implicit expiry guard used by `Fund`, `Activate`, and `CompleteMilestone` only marks the entry `expired` and ignores persistence errors. That can leave locked funds unrecovered while durable state claims the escrow expired.

## What Changes

- Make implicit expiry use the same refund semantics as explicit `Expire`.
- Propagate store update failures instead of discarding them.
- Add regression coverage for funded/active implicit expiry refund and update-error propagation.
- Reconcile dangling pending cleanup with the `ExpiresAt` contract so detectors observe old pending escrows without forcing early expiry.
- Update OpenSpec requirements so expiry behavior is fail-closed and consistent across paths.

## Impact

- Affected code: `internal/economy/escrow/engine.go`
- Affected code: `internal/economy/escrow/hub/dangling_detector.go`
- Affected tests: `internal/economy/escrow/engine_test.go`, `internal/economy/escrow/hub/dangling_detector_test.go`
- Affected specs: `economy-escrow`, `onchain-escrow`, `economy-wiring`, `p2p-documentation`, `test-coverage`
- Downstream docs: `docs/features/economy.md`
