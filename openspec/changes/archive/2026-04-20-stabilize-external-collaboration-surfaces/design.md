## Context

The external-collaboration audit identified four concrete families of drift:

- identity / trust / reputation
- pricing / negotiation / settlement
- team formation / role coordination
- workspace / shared artifacts

The underlying runtime is already substantial. The problem is not lack of capability but a control-plane mismatch between:

- what the runtime actually does,
- what operator-facing docs claim it does,
- what CLI commands imply is live,
- and which surfaces should be treated as public quote surfaces versus local policy engines.

This design captures the minimum set of technical decisions needed to make the existing runtime understandable and trustworthy without redesigning the entire P2P stack.

## Goals / Non-Goals

**Goals:**

- Make identity surfaces truthful across CLI, REST, and docs.
- Canonicalize payment-side trust defaults to one inclusive threshold.
- Clarify the split between provider-side quote surfaces and local economy policy surfaces.
- Make team/workspace/git operator surfaces honest about which paths are direct, server-backed, or tool-backed.
- Preserve the current runtime architecture unless a change directly removes a documented inconsistency.

**Non-Goals:**

- Introduce full live team HTTP APIs.
- Introduce full live workspace HTTP APIs.
- Redesign the entire trust or reputation model.
- Merge provenance, git bundle, and workspace messaging into a new artifact system.
- Replace the economy subsystem or remove existing P2P market components.

## Decisions

### Decision: Treat identity lookup as a read-only surface

The CLI and route-level identity surfaces should report the active DID when available, but they should not create or mutate identity state just to answer a query.

Why:
- A query command should not perform bundle creation or identity rotation.
- The operator-facing goal is truth-alignment, not identity lifecycle management.

Alternative considered:
- Reconstruct the full runtime identity provider in the CLI and allow it to generate missing bundle state.
Why rejected:
- It introduces side effects into a read-only command and duplicates runtime wiring.

### Decision: Canonicalize payment-side defaults through one shared constant

The post-pay threshold should come from one shared payment-side constant used by paygate and team payment negotiation.

Why:
- The audit found numeric drift (`0.7` vs `0.8`) and semantic drift (`>` vs `>=`).
- One shared constant plus one inclusive threshold rule is the simplest way to stop that drift from recurring.

Alternative considered:
- Keep separate defaults for paygate and team negotiation.
Why rejected:
- It preserves the exact ambiguity this change is supposed to remove.

### Decision: Keep public quote surfaces separate from local policy engines

`p2p.pricing` remains the provider-side public quote surface. `economy.pricing`, `economy.negotiation`, and `economy.escrow` remain local policy/engine surfaces layered above the market path.

Why:
- The runtime already has both sets of components.
- The real problem is operator confusion, not an architectural need to delete one layer immediately.

Alternative considered:
- Collapse all pricing and negotiation into one unified subsystem now.
Why rejected:
- That is a larger consolidation/refactor task than this stabilization slice.

### Decision: Treat team/workspace/git CLI as truth-aligned guidance until live operator control exists

The current CLI should tell the truth about the runtime instead of pretending to provide direct control for surfaces that are currently server-backed or tool-backed.

Why:
- The runtime subsystems exist, but direct operator control is incomplete.
- Guidance-oriented messaging is safer than pretending commands are fully live.

Alternative considered:
- Leave the optimistic examples in place until live APIs arrive.
Why rejected:
- It preserves operator-facing falsehoods and invalidates audit results.

## Risks / Trade-offs

- **Read-only identity lookup still duplicates a minimal wallet-provider selection path** → Keep it narrow, keep it read-only, and test fallback semantics directly.
- **Public-vs-policy documentation split may feel verbose** → Accept the extra wording because it prevents operator misunderstanding.
- **Guidance-oriented CLI messaging may feel less impressive than optimistic examples** → Prefer truthful operator surfaces over aspirational ones.
- **Consolidation is intentionally incomplete in this slice** → Capture the remaining larger structural work as follow-on changes rather than overloading this one.

## Migration Plan

1. Update identity surfaces first so operators can inspect the current P2P identity model reliably.
2. Canonicalize payment-side trust defaults and document the distinction from admission trust.
3. Clarify pricing, negotiation, and settlement ownership between P2P and economy surfaces.
4. Truth-align team/workspace/git operator surfaces to the actual server-backed and tool-backed model.
5. Update the audit ledger to record what landed and re-run repository verification.

Rollback strategy:
- Each slice is documentation/help/default focused and can be reverted by normal git rollback if operator messaging or threshold semantics need to be reconsidered.

## Open Questions

- Should the next change add real live team/workspace HTTP APIs or continue operator-surface truth alignment first?
- Should provenance bundle exchange remain a separate operator concept, or should it become part of a larger shared-artifact model in a later Phase 4 change?
- Once the control-plane split is clear, which pieces of `economy` should be formally merged into the P2P-facing operator surface, if any?
