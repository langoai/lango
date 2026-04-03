## Why

Phase 1 (Trust Baseline) hardened P2P security controls. Phase 2 verifies that the system's core product paths actually work end-to-end. Several critical areas had zero or insufficient test coverage: the gateway's session contract, payment tool handlers, observability event wiring, and health route semantics. Existing coverage for middleware chain order (Unit 11), orchestrator routing (Unit 12), and app wiring invariants (Unit 13) was already sufficient and confirmed via code exploration.

## What Changes

- **Gateway (Unit 8)**: Add session key contract tests and WebSocket disconnect resilience to `server_test.go`. Remove empty `gateway_test.go`.
- **Payment tools (Unit 10)**: Add `tools_test.go` covering base/conditional tool set, safety levels, and parameter validation.
- **EventBus contracts (Unit 14)**: Add `eventbus_contracts_test.go` verifying that published events produce expected collector state changes.
- **Health routes (Unit 15)**: Add `routes_observability_test.go` testing `/health/detailed` and `/metrics` route semantics at the route-registration level.

Units 9, 11, 12, 13 were confirmed as already covered by existing tests (`parity_test.go`, `chain_order_test.go`, `orchestrator_test.go`) and skipped.

## Capabilities

### New Capabilities

(none — test-only change, no new runtime capabilities)

### Modified Capabilities

(none — no spec-level behavior changes)

## Impact

- **Code**: `internal/gateway/server_test.go` (modified), `internal/tools/payment/tools_test.go` (new), `internal/app/eventbus_contracts_test.go` (new), `internal/app/routes_observability_test.go` (new)
- **Removed**: `internal/gateway/gateway_test.go` (empty file)
- **Behavior**: No runtime behavior changes — test-only
