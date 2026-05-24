## Context

The codebase already moved bootstrap ownership and runtime readers behind broker/storage capabilities, but payment mutators still rely on direct parent-side Ent access. Both CLI and app wiring open a session store, assert `*session.EntStore`, and extract `Client()` to build spending limiters, payment services, and settlement services. That preserves a raw ORM escape hatch in production paths even after `PaymentClient()` was removed from the storage facade.

Payment send flow and P2P settlement share the same `PaymentTx` persistence concerns: create a record, update status transitions, and query recent usage for spending limits. The persistence contract is narrower than a full Ent client, so the next slice should expose that narrower contract explicitly and let service construction stay in payment/app layers.

## Goals / Non-Goals

**Goals:**
- Remove direct `*ent.Client` extraction from payment CLI and app payment/settlement wiring.
- Introduce explicit payment transaction persistence interfaces that are small enough to serve both payment send flow and P2P settlement.
- Keep spending-limit enforcement and payment service construction working with the new storage-facing capabilities.
- Preserve current CLI behavior and payment transaction semantics.

**Non-Goals:**
- Broker-enable all payment mutators in runtime-heavy paths during this slice.
- Migrate knowledge, learning, inquiry, or agent memory mutators in this slice.
- Change payment command UX or on-chain behavior.
- Introduce new blockchain providers or payment protocols.

## Decisions

### 1. Split payment persistence from service construction

Payment service construction remains in `internal/payment` and app/CLI wiring, but payment transaction persistence moves behind explicit interfaces. This keeps storage ownership narrow without turning the storage facade into a giant payment service factory.

Alternatives considered:
- Add `NewPaymentService(...)` directly to the storage facade: rejected because it would recreate a broad service-factory escape hatch.
- Keep extracting `*ent.Client` from `session.EntStore`: rejected because it defeats the broker-boundary objective.

### 2. Use one narrow transaction store contract for both payment and settlement

`payment.Service` and `p2p/settlement.Service` both need to create and update `PaymentTx` records. A shared `payment.TxStore` contract lets both services use the same storage-facing capability without depending on Ent.

Alternatives considered:
- Separate store interfaces for payment and settlement: rejected because the data model and write lifecycle are the same.
- Move settlement-specific persistence into `p2p/settlement`: rejected because `PaymentTx` is a payment-domain entity.

### 3. Keep spending-limit enforcement behind a usage-reader contract

`wallet.EntSpendingLimiter` currently queries Ent directly. This slice replaces that dependency with a narrower usage-reader interface so spending limits can be constructed through storage capabilities without a raw ORM handle.

Alternatives considered:
- Fold spending checks into `payment.Service`: rejected because the limiter is reused by X402 interception and other payment call sites.
- Query usage through ad hoc storage facade methods only in CLI/app: rejected because it duplicates limit logic and weakens testability.

## Risks / Trade-offs

- **[Risk] Payment service constructor changes ripple into tests** → Mitigation: add Ent-backed adapter constructors in `payment`/`wallet` packages and update existing tests to use those adapters instead of raw clients.
- **[Risk] Settlement persistence semantics diverge from send flow** → Mitigation: share one transaction store contract and keep status-transition helpers centralized.
- **[Risk] Broker mode still does not run payment mutators end-to-end** → Mitigation: keep this slice focused on removing raw parent-side ORM access first; broker-backed mutator RPCs can build on the same explicit store contracts next.
- **[Risk] Hidden raw client usage remains in production code** → Mitigation: search/arch-test-driven verification before closing the change.

## Migration Plan

1. Add explicit payment transaction store and spending-usage interfaces plus Ent-backed adapters.
2. Rewire payment service and settlement service to use those interfaces.
3. Rewire CLI payment init and app payment/P2P wiring to consume storage-facing capabilities only.
4. Update OpenSpec tasks/specs, then run build/test/validate.

## Open Questions

- Whether the next slice should expose broker-backed payment mutator RPCs directly or first move protected knowledge/learning/inquiry mutators behind similar interfaces.
