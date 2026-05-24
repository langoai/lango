## Context

The product now has:

- source-primary exportability
- structured artifact release approval
- lite dispute-ready receipts

What remains missing is the front-end approval model for deciding whether a transaction may start with an upfront payment. This decision should produce durable evidence and transaction-level canonical state, but it should not yet execute payments or escrow logic.

## Goals / Non-Goals

**Goals:**

- Add a dedicated upfront payment approval domain model.
- Emit structured `approve / reject / escalate` outcomes with payment mode recommendation.
- Update transaction receipts with canonical payment approval state.
- Add a minimal agent-facing prepayment approval tool and truthful docs.

**Non-Goals:**

- Actual payment execution gating.
- Escrow execution.
- Human approval UI.
- Full transaction orchestration.
- Payment dispute adjudication.

## Decisions

### 1. Treat upfront payment approval as a subtype of the approval receipt family

This keeps the slice aligned with the artifact release approval work and avoids creating an unrelated payment-only evidence system.

Alternative considered:

- audit-only storage
  - rejected because it weakens linkage to transaction canonical state

### 2. Update transaction receipts directly

The first slice must do more than log a decision. It should update transaction-level canonical payment approval state so later surfaces have one current truth.

Alternative considered:

- leave transaction receipt untouched and rely on audit rows
  - rejected because it spreads canonical state across multiple places

### 3. Keep payment mode as recommendation only

The first slice may recommend `prepay`, `escrow`, `escalate`, or `reject`, but it does not execute those paths.

Alternative considered:

- combine decision and execution
  - rejected as too large for the first slice

## Risks / Trade-offs

- **[Risk]** The term “approval” may be interpreted as execution permission rather than decisioning.
  - **Mitigation:** Keep docs explicit that this slice stops at structured decisioning and receipts.

- **[Risk]** Suggested payment mode may look more concrete than it really is.
  - **Mitigation:** State clearly that the mode is recommendation only in this slice.

- **[Risk]** Transaction receipt updates may be too eager if later execution differs.
  - **Mitigation:** Scope the canonical update specifically to payment approval state, not to final settlement execution.
