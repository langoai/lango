## Context

Phase 2 of the comprehensive roadmap targets whole-product verification. Code exploration revealed that several units (11, 12, 13) already had sufficient test coverage, while others (8, 10, 14, 15) had specific gaps. This design focuses on the 4 units that needed work.

## Goals / Non-Goals

**Goals:**
- Verify gateway session key assignment contracts
- Verify payment tool handler metadata, conditional wiring, and parameter validation
- Verify observability event contracts (event → collector state)
- Verify health/metrics route registration semantics

**Non-Goals:**
- Adding runtime features or changing behavior
- Full E2E tests requiring live LLM providers
- Exhaustive unit testing of every tool handler

## Decisions

1. **Gateway tests in `server_test.go`, not new file** — Existing test patterns (newTestServer, Client struct usage) are all in server_test.go. Adding cases there is more consistent than creating a separate integration test file. The empty `gateway_test.go` is removed.

2. **Payment tests use real types, not mocks** — `SecretsStore` and `Interceptor` are concrete structs. Tests use zero-value pointers or minimal constructors to test `BuildTools()` conditional logic, then verify tool metadata without calling handlers (except parameter validation).

3. **EventBus contracts replicate wiring, don't test through app** — Tests create a standalone Bus + Collector and wire the same subscriptions as `wiring_observability.go`. This avoids full `app.New()` bootstrap while testing the exact same event contract.

4. **Health route tests at registration level** — Tests call `registerObservabilityRoutes()` directly on a chi.Router with `httptest.Server`, avoiding full app bootstrap. This tests route semantics in isolation.

## Risks / Trade-offs

- **[Trade-off] EventBus tests duplicate subscription wiring** → Acceptable: the alternative (full app bootstrap) would make tests slow and fragile. The duplication is minimal (3 lines).
- **[Trade-off] Skipped units (9, 11, 12, 13)** → `parity_test.go` already covers what Unit 9 and 13 would test. Confirmed via code exploration.
