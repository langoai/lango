## Context

Earlier broker-ownership changes moved bootstrap, readers, and payment mutator setup onto narrower storage-facing capabilities. The remaining risk is regression: removed raw accessors can be reintroduced later, or payment wiring can quietly fall back to `session.EntStore.Client()` again. Broker transport correctness is also critical enough that its regression tests should be treated as part of the enforcement boundary, not just incidental unit coverage.

## Goals / Non-Goals

**Goals:**
- Mechanically prevent removed storage facade escape hatches from reappearing in production packages.
- Mechanically prevent payment production wiring from extracting Ent clients directly.
- Treat broker transport regression tests as part of the enforced boundary.
- Remove the last payment wiring fallback that reconstructs a tx store from `session.EntStore.Client()`.

**Non-Goals:**
- Eliminate every remaining `session.EntStore` type assertion in the whole app.
- Move knowledge/learning/inquiry/agent-memory mutators onto broker RPCs in this change.
- Change CLI UX or transport protocol shape.

## Decisions

### 1. Enforce boundary with archtest, not comments or conventions

Boundary rules are encoded as `internal/archtest` grep-based tests scoped to production packages. This keeps the rule cheap to run under `go test ./...` and consistent with existing bootstrap boundary tests.

### 2. Scope production enforcement to concrete forbidden patterns

This slice forbids:
- removed storage facade escape hatches (`EntClient`, `RawDB`, `FTSDB`, `PaymentClient`)
- `storage.WithEntClient` / `storage.WithRawDB` in production packages
- `entStore.Client()` reintroduction in payment production wiring

It does not try to outlaw every `session.EntStore` use yet because that would exceed the currently migrated ownership surface.

### 3. Treat existing broker transport tests as release gates

The code already has regression tests for concurrent writes, large responses, and graceful shutdown. This change keeps them explicit in the change contract and verification checklist, rather than inventing a second enforcement mechanism.

## Risks / Trade-offs

- **[Risk] Archtests become too broad and block valid refactors** → Mitigation: scope package patterns tightly and document allowed surfaces.
- **[Risk] Payment wiring loses fallback flexibility** → Mitigation: rely on `bootstrap.Result.Storage`, which is already the canonical runtime dependency hub.
- **[Risk] Transport gate relies on conventional test names** → Mitigation: keep those tests in `internal/storagebroker/protocol_test.go` and run them as part of the required verification.
