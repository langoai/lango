# Pricing Negotiation Settlement Audit

## Purpose

This document is the next detailed audit ledger under the Lango master document.

It exists to review the pricing, negotiation, settlement, and escrow boundary that `knowledge exchange v1` already depends on:

- provider-side public quotes,
- local pricing and negotiation policy,
- direct payment settlement controls,
- escrow-backed exchange progression.

The purpose of this audit is not to design the full runtime end-to-end transaction model yet. Its purpose is to lock the current control-plane relationship against the real code and documentation surface so later runtime, settlement, and escrow work inherits one consistent interpretation.

## Relationship to the Master Document

This audit sits underneath `docs/architecture/master-document.md` and must use that document's constitution, capability taxonomy, audit vocabulary, and track-routing rules.

It does not redefine what Lango is, replace the product constitution, or create new top-level capability areas or execution tracks. Its role is to apply the master document's framework to the `pricing / negotiation / settlement` surface in detailed ledger form for the `P2P Knowledge Exchange Track`.

## Document Ownership

- Primary capability area: `External Collaboration & Economic Exchange`
- Primary execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas:
  - `Trust, Security & Policy`
  - `Execution, Continuity & Accountability`
- Secondary tracks:
  - `Stabilization Track`

## Audit Order

1. Pricing Surface
2. Negotiation
3. Settlement
4. Escrow

## Audit Method

This ledger adopts the master document's minimum audit schema for detailed row-level work.

Each row must include:

- feature name,
- capability area,
- product-path linkage,
- current surface area,
- core value,
- current problem,
- judgment,
- execution track,
- secondary capability areas,
- secondary tracks.

Each feature family is judged against:

- current operator-facing clarity,
- fit with the `knowledge exchange v1` boundary,
- consistency between docs and runtime behavior,
- whether the surface is public quote, local policy, settlement control, or escrow progression,
- whether the current capability needs removal, merging, or stabilization.

Allowed judgments:

- `keep`
- `stabilize`
- `merge`
- `defer`
- `remove`

The judgment baseline for this audit is deliberately narrow:

- Does the surface create a coherent early external exchange control plane?
- Does it preserve the distinction between provider-side quotes, local policy engines, settlement, and escrow?
- Does it describe off-chain accrual and postpay as trust-conditional Phase 2 capability rather than a fully general default?

## Current Surface Map

| Feature family | Primary phase | Current surface clues | Audit status |
| --- | --- | --- | --- |
| Pricing Surface | Phase 1-2 | `docs/features/p2p-network.md`, `docs/features/economy.md`, `docs/cli/p2p.md`, `internal/cli/p2p/pricing.go`, `internal/economy/pricing/*`, `internal/app/p2p_routes.go` | Detailed audit complete (`stabilize`) |
| Negotiation | Phase 1-2 | `docs/features/economy.md`, `internal/economy/negotiation/*`, `internal/economy/tools.go`, `internal/app/wiring_economy.go` | Detailed audit complete (`stabilize`) |
| Settlement | Phase 1-2 | `docs/security/upfront-payment-approval.md`, `docs/security/actual-payment-execution-gating.md`, `internal/paymentapproval/*`, `internal/paymentgate/*`, `internal/tools/payment/*`, `internal/app/tools_p2p.go` | Detailed audit complete (`stabilize`) |
| Escrow | Phase 1-2 | `docs/security/escrow-execution.md`, `docs/features/economy.md`, `internal/economy/escrow/*`, `internal/escrowexecution/*`, `internal/app/tools_escrow.go` | Detailed audit complete (`stabilize`) |

## Baseline Control-Plane Model

The following control-plane model is locked for this audit and should remain stable for later runtime and product-path design work.

- `p2p.pricing` is the provider-side public quote surface exposed to remote peers.
- `economy.pricing` is the local pricing and policy engine that may influence market behavior, but it is not the same public quote surface.
- negotiation is real, but it is still under-surfaced in the operator-facing `knowledge exchange v1` transaction story.
- settlement and escrow are distinct rows because the current approval and direct-payment path is not the same thing as the escrow lifecycle.
- off-chain accrual and postpay remain Phase 2, trust-conditional, and still limited rather than the fully generalized default.

This means the current model is intentionally layered:

- public quote exposure lives under `p2p.pricing`,
- local pricing and negotiation policy lives under `economy.*`,
- settlement currently means approval plus direct-payment execution control and post-pay progression semantics,
- escrow currently means recommendation, bound execution input, and the first `create + fund` lifecycle slice.

## Detailed Audit: Pricing Surface

### Audit Record

- Feature name: `Pricing Surface`
- Capability area: `External Collaboration & Economic Exchange`
- Product-path linkage: `Phase 1: Knowledge Exchange`, `Phase 2: Result Exchange with Controlled Execution`
- Current surface area: `docs/features/p2p-network.md`, `docs/features/economy.md`, `docs/cli/p2p.md`, `internal/cli/p2p/pricing.go`, `internal/app/p2p_routes.go`, `internal/economy/pricing/*`, `internal/app/wiring_economy.go`
- Core value: `Expose a provider-side quote surface to remote peers while preserving a separate local policy engine for trust-sensitive pricing decisions.`
- Current problem: `The pricing capability is real, but operators still have to infer the difference between public quote exposure and local dynamic pricing policy by reading across P2P and economy surfaces.`
- Judgment: `stabilize`
- Execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas:
  - `Trust, Security & Policy`
- Secondary tracks:
  - `Stabilization Track`

### Findings

1. `Major` The public quote surface is already real and operator-visible.
   - `docs/features/p2p-network.md`, `docs/cli/p2p.md`, `internal/cli/p2p/pricing.go`, and `internal/app/p2p_routes.go` all expose the same narrow shape: `perQuery`, tool-specific overrides, and USDC-denominated quotes.
   - That makes `p2p.pricing` a real provider-side quote surface rather than a placeholder.

2. `Major` `economy.pricing` is also real, but it serves a different role.
   - `docs/features/economy.md` explicitly describes the economy subsystem as the local policy layer and says it is not the same thing as `p2p.pricing`.
   - `internal/economy/pricing/engine.go` computes trust-sensitive and rule-based quotes using reputation and a minimum price floor.
   - `internal/app/wiring_economy.go` can wire that engine into the payment path when enabled.

3. `Major` The current operator model still under-explains the relationship between the two surfaces.
   - The public surface shows static provider quotes.
   - The local engine can compute peer-sensitive pricing.
   - That is a legitimate layered design, but it still needs one canonical explanation so operators do not treat the two surfaces as duplicate public APIs.

### Assessment

- `Pricing Surface` should be kept.
- The correct action is `stabilize`:
  - keep `p2p.pricing` as the provider-side public quote surface,
  - keep `economy.pricing` as the local pricing and policy engine,
  - document the two as complementary rather than competing control planes.

## Detailed Audit: Negotiation

### Audit Record

- Feature name: `Negotiation`
- Capability area: `External Collaboration & Economic Exchange`
- Product-path linkage: `Phase 1: Knowledge Exchange`, `Phase 2: Result Exchange with Controlled Execution`
- Current surface area: `docs/features/economy.md`, `internal/economy/negotiation/*`, `internal/economy/tools.go`, `internal/app/wiring_economy.go`, `internal/p2p/protocol/*`
- Core value: `Allow peers to move beyond static quotes through multi-round price negotiation before work or settlement proceeds.`
- Current problem: `Negotiation exists in the runtime and in economy tools, but it is still under-surfaced as part of the main operator-facing knowledge-exchange transaction story.`
- Judgment: `stabilize`
- Execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas:
  - `Trust, Security & Policy`
- Secondary tracks:
  - `Stabilization Track`

### Findings

1. `Major` Negotiation is already a real engine-level capability.
   - `docs/features/economy.md` documents a real negotiation lifecycle with sessions, rounds, timeout, and completion or rejection states.
   - `internal/economy/tools.go` exposes `economy_negotiate` and `economy_negotiate_status` as operator entry points.

2. `Major` The runtime wiring is broader than the current public story suggests.
   - `internal/app/wiring_economy.go` initializes the negotiation engine, wires pricing into auto-response, publishes negotiation events, and attaches a negotiator to the P2P protocol handler.
   - That means negotiation is not just a local helper. It is part of the actual runtime path when enabled.

3. `Major` Negotiation is still under-surfaced relative to pricing and direct payment.
   - The P2P-facing docs primarily read as quote, pay, then invoke.
   - Negotiation is documented more as an economy subsystem feature than as a canonical exchange-stage in the public transaction flow.
   - The current issue is therefore surfacing and articulation, not capability absence.

### Assessment

- `Negotiation` should be kept.
- The correct action is `stabilize`:
  - preserve negotiation as a real runtime capability,
  - keep it distinct from both static quoting and settlement,
  - strengthen the operator-facing articulation of when negotiation participates in the transaction path.

## Detailed Audit: Settlement

### Audit Record

- Feature name: `Settlement`
- Capability area: `External Collaboration & Economic Exchange`
- Product-path linkage: `Phase 1: Knowledge Exchange`, `Phase 2: Result Exchange with Controlled Execution`
- Current surface area: `docs/features/p2p-network.md`, `docs/security/upfront-payment-approval.md`, `docs/security/actual-payment-execution-gating.md`, `internal/paymentapproval/*`, `internal/paymentgate/*`, `internal/tools/payment/*`, `internal/app/tools_p2p.go`, `internal/p2p/paygate/*`, `internal/p2p/settlement/*`
- Core value: `Turn paid exchange into a controlled execution path by linking approval state, direct payment execution, a real on-chain settlement runtime, and limited deferred-settlement behavior.`
- Current problem: `Settlement is real, but the current model is still split across approval, direct-payment gating, the on-chain settlement service, and trust-conditional deferred-payment hooks, so the end-to-end progression is not yet one fully consolidated operator story.`
- Judgment: `stabilize`
- Execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas:
  - `Trust, Security & Policy`
  - `Execution, Continuity & Accountability`
- Secondary tracks:
  - `Stabilization Track`

### Findings

1. `Major` Settlement already has a real receipt-backed approval and execution-control path.
   - `docs/security/upfront-payment-approval.md` defines a first slice that records canonical payment approval state and settlement hints.
   - `docs/security/actual-payment-execution-gating.md` defines a fail-closed direct-payment gate for `payment_send` and `p2p_pay`.
   - `internal/paymentgate/service.go`, `internal/tools/payment/payment.go`, and `internal/app/tools_p2p.go` enforce that direct payment executes only when the receipt state is approved for `prepay`.

2. `Major` Settlement also includes a real on-chain runtime surface, not only approval plus direct-payment gating.
   - `docs/features/p2p-network.md` describes a settlement service that handles asynchronous USDC settlement for paid P2P tool invocations.
   - `internal/p2p/settlement/service.go` subscribes to paid-execution events, builds `transferWithAuthorization` transactions, submits them on-chain with retry, waits for confirmation, and records success or failure for reputation tracking.
   - This makes settlement a real runtime path, even though the overall operator story is still fragmented.

3. `Major` The P2P payment path still includes trust-conditional post-pay semantics, but they should be read conservatively.
   - `docs/features/p2p-network.md` and `internal/p2p/paygate/*` keep `postPayMinScore` as the payment-side threshold for whether execution may happen before settlement.
   - The current high-trust path is closer to deferred ledger and event-hook behavior than to a fully mature, generalized post-pay settlement runtime.
   - This row should therefore treat post-pay and off-chain accrual as limited Phase 2 behavior, not as a broad landed default.

4. `Major` Settlement should remain distinct from escrow.
   - The current direct-payment gate decides whether a `prepay` execution may proceed now.
   - That is a different maturity slice from escrow recommendation and escrow execution.
   - Treating them as one row would blur two different progression models that now have separate evidence and runtime paths.

### Assessment

- `Settlement` should be kept.
- The correct action is `stabilize`:
  - keep approval state, direct-payment gating, the on-chain settlement service, and trust-conditional deferred-payment behavior inside the settlement row,
  - keep off-chain accrual and postpay explicitly limited, Phase 2, and less mature than the direct-payment path,
  - avoid collapsing settlement into escrow.

## Detailed Audit: Escrow

### Audit Record

- Feature name: `Escrow`
- Capability area: `External Collaboration & Economic Exchange`
- Product-path linkage: `Phase 1: Knowledge Exchange`, `Phase 2: Result Exchange with Controlled Execution`
- Current surface area: `docs/features/economy.md`, `docs/security/escrow-execution.md`, `internal/economy/escrow/*`, `internal/escrowexecution/*`, `internal/app/tools_escrow.go`, `internal/app/tools_meta.go`, `internal/app/wiring_economy.go`
- Core value: `Provide a higher-friction protected exchange path for paid work by binding approved escrow intent to a real runtime execution and longer lifecycle.`
- Current problem: `Escrow is now a real engine and execution slice, but the current knowledge-exchange path only lands recommendation plus `create + fund`, leaving activation, release, refund, and dispute as explicit follow-on lifecycle work.`
- Judgment: `stabilize`
- Execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas:
  - `Trust, Security & Policy`
  - `Execution, Continuity & Accountability`
- Secondary tracks:
  - `Stabilization Track`

### Findings

1. `Major` Escrow is already a real engine-level subsystem with its own lifecycle.
   - `docs/features/economy.md` documents create, fund, activate, submit-work, release, refund, dispute, and resolve states.
   - `internal/app/tools_escrow.go` exposes concrete escrow tools for those lifecycle actions.
   - `internal/app/wiring_economy.go` initializes the escrow engine and selects settlement mode when the economy subsystem is enabled.

2. `Major` The `knowledge exchange v1` path now includes a real first escrow execution slice.
   - `docs/security/upfront-payment-approval.md` binds escrow execution input when approval recommends `escrow`.
   - `docs/security/escrow-execution.md` and `internal/escrowexecution/service.go` execute the bound recommendation through a receipt-backed `create + fund` path.
   - `internal/app/tools_meta.go` exposes `execute_escrow_recommendation` as the current operator entry point.

3. `Major` The landed escrow slice is still intentionally incomplete.
   - The current execution status stops at `pending`, `created`, `funded`, or `failed`.
   - `docs/security/escrow-execution.md` explicitly leaves activation, milestone release, refund, and dispute handling as follow-on work.
   - That means escrow is no longer merely planned, but it is still only partially surfaced for the current product path.

4. `Major` Escrow should stay distinct from settlement.
   - Settlement already has its own approval and direct-payment gate story.
   - Escrow has a separate recommendation, bound execution input, and lifecycle completion problem.
   - Keeping these as separate rows preserves the real current product distinction.

### Assessment

- `Escrow` should be kept.
- The correct action is `stabilize`:
  - keep escrow as a distinct row from settlement,
  - preserve the landed `create + fund` execution slice,
  - treat lifecycle completion as follow-on work rather than over-claiming current maturity.

## Assessment

All four rows remain `stabilize`: the capability family is real, but the control-plane and progression model still need consolidation.

The key lock for follow-on work is:

- `p2p.pricing` is the provider-side public quote surface,
- `economy.pricing` is the local pricing and policy engine,
- negotiation is real but under-surfaced,
- settlement and escrow are distinct rows,
- off-chain accrual and postpay are Phase 2, trust-conditional, and still limited.

## Follow-On Design Inputs

1. `knowledge exchange runtime` end-to-end design
2. settlement follow-on work
3. escrow lifecycle completion
