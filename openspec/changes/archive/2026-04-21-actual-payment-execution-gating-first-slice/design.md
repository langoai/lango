## Context

The product now has multiple decision layers:

- exportability
- artifact release approval
- dispute-ready receipts
- upfront payment approval

What is still missing is the final execution control layer that decides whether a direct payment may actually proceed. This slice must stay narrow: it should gate only `payment_send` and `p2p_pay`, and only for direct payment execution.

## Goals / Non-Goals

**Goals:**

- Add a shared direct payment execution gate service.
- Wire the gate into `payment_send` and `p2p_pay`.
- Use receipt-backed canonical payment approval state as the primary source of truth.
- Emit allow/deny execution evidence into audit and receipt trails.

**Non-Goals:**

- Escrow execution gating.
- Human payment approval UI.
- Middleware-wide enforcement for every payment-like tool.
- Full transaction orchestration.

## Decisions

### 1. Use a shared gate service with handler-local integration

The handlers remain the integration point, but the decision logic lives in one shared gate service. This avoids drift between `payment_send` and `p2p_pay` while keeping the first slice small.

Alternative considered:

- inline checks inside each handler
  - rejected because duplicated logic and event recording would drift quickly

### 2. Use transaction receipt as the primary key and source of truth

The gate reads the linked transaction receipt first, then follows linked submission or receipt detail as needed. This matches the canonical-state-first design of the receipts model.

Alternative considered:

- approval receipt only
  - rejected because execution should consume canonicalized state, not reconstruct it

### 3. Deny when transaction receipt id is missing

This first slice is explicitly receipt-backed. Direct payment execution without `transaction_receipt_id` is denied rather than silently falling back to legacy behavior.

Alternative considered:

- allow legacy untracked execution
  - rejected because it would weaken the point of the new gate

## Risks / Trade-offs

- **[Risk]** Requiring receipt IDs may temporarily block older direct payment paths.
  - **Mitigation:** Keep the docs explicit that the first gate slice applies only to receipt-backed direct payment execution.

- **[Risk]** Direct payment only may be confused with broader settlement orchestration.
  - **Mitigation:** State clearly that escrow execution is still out of scope.

- **[Risk]** A narrow shared gate service may later need refactoring for broader middleware enforcement.
  - **Mitigation:** Keep the service small and focused on direct payment execution only.
