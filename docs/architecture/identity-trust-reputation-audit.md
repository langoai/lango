# Identity Trust Reputation Audit

## Purpose

This document is the next detailed audit ledger under the Lango master document.

It exists to review the identity, trust, and reputation boundary that `knowledge exchange v1` already depends on:

- identity continuity,
- trust entry,
- reputation,
- revocation and trust decay.

The purpose of this audit is not to design `reputation v2` yet. Its purpose is to lock the current relationship model against the real code and documentation surface so later pricing, runtime, and settlement work inherits one consistent interpretation.

## Relationship to the Master Document

This audit sits underneath `docs/architecture/master-document.md` and must use that document's constitution, capability taxonomy, audit vocabulary, and track-routing rules.

It does not redefine what Lango is, replace the product constitution, or create new top-level capability areas or execution tracks. Its role is to apply the master document's framework to the identity/trust/reputation surface in detailed ledger form for the `P2P Knowledge Exchange Track`.

## Document Ownership

- Primary capability area: `External Collaboration & Economic Exchange`
- Primary execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas:
  - `Trust, Security & Policy`
  - `Execution, Continuity & Accountability`
- Secondary tracks:
  - `Stabilization Track`

## Audit Order

1. Identity Continuity
2. Trust Entry
3. Reputation
4. Revocation & Trust Decay

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
- separation of bootstrap trust, earned trust, and payment friction,
- whether the current surface needs removal, merging, or stabilization.

Allowed judgments:

- `keep`
- `stabilize`
- `merge`
- `defer`
- `remove`

The judgment baseline for this audit is deliberately narrow:

- Does the surface create a coherent early external exchange boundary?
- Does it preserve the distinction between root trust, agent reputation, and transaction trust?
- Does it express runtime safety controls without over-claiming durable reputation semantics?

## Current Surface Map

| Feature family | Primary phase | Current surface clues | Audit status |
| --- | --- | --- | --- |
| Identity Continuity | Phase 1-2 | `docs/features/p2p-network.md`, `docs/cli/p2p.md`, `docs/gateway/http-api.md`, `internal/p2p/identity/*`, `internal/p2p/handshake/*`, `internal/app/wiring_p2p.go` | Detailed audit complete (`stabilize`) |
| Trust Entry | Phase 1-2 | `docs/security/authentication.md`, `docs/features/p2p-network.md`, `internal/gateway/auth.go`, `internal/p2p/firewall/*`, `internal/p2p/handshake/security_events.go`, `internal/p2p/paygate/*` | Detailed audit complete (`stabilize`) |
| Reputation | Phase 1-2 | `docs/features/p2p-network.md`, `docs/cli/p2p.md`, `internal/p2p/reputation/*`, `internal/app/p2p_routes.go`, `internal/p2p/team/payment.go` | Detailed audit complete (`stabilize`) |
| Revocation & Trust Decay | Phase 1-2 | `docs/features/p2p-network.md`, `internal/p2p/handshake/security_events.go`, `internal/p2p/discovery/gossip.go`, `internal/p2p/reputation/*` | Detailed audit complete (`stabilize`) |

## Baseline Relationship Model

The following relationship model is locked for this audit and should remain stable for later runtime and market-facing design work.

- `owner-root trust` provides a bootstrap ceiling and floor, but it does not grant a new agent full inherited trust.
- `agent/domain reputation` is earned from actual exchange history, fulfillment, and repeated collaboration outcomes.
- `admission trust` and `payment trust` are separate gates. The same inputs may influence both, but they answer different product questions.
- `operational signals` and `durable negative reputation` are separate concepts. Immediate runtime safety actions may rely on operational signals; durable reputation should require stronger adjudication.
- `bootstrap trust` and `earned trust` are separate states. New agents begin under constrained trust conditions even when the owner is already trusted.

This means the current model is intentionally mixed:

- owner identity and root accountability establish continuity,
- peer and agent history establish earned reputation,
- admission controls decide whether a boundary crossing is allowed,
- payment controls decide what friction applies after an exchange path exists,
- runtime safety events may revoke access faster than they should permanently damage durable reputation.

## Detailed Audit: Identity Continuity

### Audit Record

- Feature name: `Identity Continuity`
- Capability area: `External Collaboration & Economic Exchange`
- Product-path linkage: `Phase 1: Knowledge Exchange`, `Phase 2: Result Exchange with Controlled Execution`
- Current surface area: `docs/features/p2p-network.md`, `docs/cli/p2p.md`, `docs/gateway/http-api.md`, `internal/p2p/identity/*`, `internal/p2p/handshake/*`, `internal/app/wiring_p2p.go`
- Core value: `Keep a remote agent legible across handshake, discovery, API, and payment flows through cryptographic continuity rather than a disposable session-only identity.`
- Current problem: `Identity continuity is real, but the operator model still has to reason across wallet-derived DID, bundle-backed DID, peer ID, and gateway-visible identity surfaces at the same time.`
- Judgment: `stabilize`
- Execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas:
  - `Trust, Security & Policy`
- Secondary tracks:
  - `Stabilization Track`

### Findings

1. `Major` Identity continuity is already a real runtime capability, not a placeholder.
   - `internal/p2p/identity/identity.go` supports both legacy `did:lango:<hex>` and bundle-backed `did:lango:v2:<hash>` forms.
   - `internal/app/wiring_p2p.go` prefers the bundle provider when identity material is available and falls back to legacy identity otherwise.
   - `docs/features/p2p-network.md`, `docs/cli/p2p.md`, and `docs/gateway/http-api.md` all expose the active DID as a user-facing surface.

2. `Major` The current surface still requires one conceptual mapping too many for operators.
   - The docs correctly describe DID continuity, but the actual runtime still exposes identity through multiple views: DID, peer ID, handshake signer, and session token.
   - That is acceptable for `knowledge exchange v1`, but it means the current capability needs consolidation rather than expansion.

3. `Major` Owner-root trust belongs here only as a continuity floor, not as a substitute for agent reputation.
   - The design baseline and the `P2P Knowledge Exchange Track` already frame the system as `owner-root trust plus agent-specific reputation`.
   - Identity continuity should therefore preserve who the owner-root is and which agent instance is speaking, while leaving earned trust to the reputation row.

### Assessment

- `Identity Continuity` should be kept.
- The correct action is `stabilize`:
  - keep the current DID and bundle-based continuity model,
  - preserve owner-root trust as the bootstrap floor for identity continuity,
  - avoid collapsing identity continuity into a claim that new agents are fully trusted.

## Detailed Audit: Trust Entry

### Audit Record

- Feature name: `Trust Entry`
- Capability area: `External Collaboration & Economic Exchange`
- Product-path linkage: `Phase 1: Knowledge Exchange`, `Phase 2: Result Exchange with Controlled Execution`
- Current surface area: `docs/security/authentication.md`, `docs/features/p2p-network.md`, `internal/gateway/auth.go`, `internal/p2p/firewall/*`, `internal/p2p/handshake/security_events.go`, `internal/p2p/paygate/*`
- Core value: `Define which peers can cross the boundary, under what conditions they keep access, and which gate is about admission versus payment.`
- Current problem: `Gateway auth, handshake approval, firewall admission, and payment friction are all real, but they still read like adjacent mechanisms instead of one operator-facing trust-entry model.`
- Judgment: `stabilize`
- Execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas:
  - `Trust, Security & Policy`
- Secondary tracks:
  - `Stabilization Track`

### Findings

1. `Major` Admission trust is already implemented as a real multi-layer gate.
   - `internal/gateway/auth.go` establishes gateway session-based auth when OIDC is configured.
   - `docs/features/p2p-network.md` describes the P2P approval pipeline as firewall ACL, reputation check, and owner approval.
   - `internal/app/p2p_routes.go` protects `/api/p2p` with gateway auth whenever auth is configured.

2. `Major` Admission trust and payment trust must stay separate in the operator model.
   - Admission trust answers whether the peer is allowed in at all.
   - Payment trust answers whether a paid invocation can use post-pay or must use upfront payment.
   - `internal/p2p/paygate/gate.go` enforces payment-tier routing after the invocation path already exists, which is a different question from handshake or firewall entry.

3. `Major` New peers begin under constrained trust even inside a trusted owner context.
   - `internal/p2p/reputation/store.go` and the P2P docs allow new peers through the admission threshold with the benefit of the doubt.
   - `internal/p2p/paygate/gate.go` still withholds post-pay unless the peer earns a score at or above the payment threshold.
   - This is the correct current expression of bootstrap trust versus earned trust.

### Assessment

- `Trust Entry` should be kept.
- The correct action is `stabilize`:
  - document gateway auth, firewall entry, and handshake/session controls as one admission model,
  - preserve the separation between admission trust and payment trust,
  - keep new peers constrained even when owner-root trust prevents a full zero-trust posture.

## Detailed Audit: Reputation

### Audit Record

- Feature name: `Reputation`
- Capability area: `External Collaboration & Economic Exchange`
- Product-path linkage: `Phase 1: Knowledge Exchange`, `Phase 2: Result Exchange with Controlled Execution`
- Current surface area: `docs/features/p2p-network.md`, `docs/cli/p2p.md`, `internal/p2p/reputation/*`, `internal/app/p2p_routes.go`, `internal/p2p/team/payment.go`
- Core value: `Capture exchange history so trust can be earned from actual outcomes instead of being inherited wholesale from owner identity or one-time entry approval.`
- Current problem: `The reputation store and operator query surfaces are real, but the durable meaning of owner-root trust, agent/domain reputation, and payment-side trust reduction is not yet frozen into one canonical model.`
- Judgment: `stabilize`
- Execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas:
  - `Trust, Security & Policy`
- Secondary tracks:
  - `Stabilization Track`

### Findings

1. `Major` Reputation already has a durable runtime store and visible operator surface.
   - `internal/p2p/reputation/store.go` persists successful exchanges, failures, timeouts, trust score, first-seen, and last-interaction state.
   - `internal/app/p2p_routes.go` exposes read-only reputation details through `/api/p2p/reputation`.
   - `docs/gateway/http-api.md` and `docs/cli/p2p.md` present reputation as a first-class operator-visible concept.

2. `Major` Reputation should remain agent/domain-earned rather than owner-inherited.
   - The current track language already distinguishes owner-root trust from agent-specific reputation.
   - The reputation store is keyed by `peerDID` and updated from actual exchange outcomes, which aligns with an earned-reputation model rather than a root-trust inheritance model.

3. `Major` Payment trust should continue to consume reputation as a policy input, not redefine reputation itself.
   - `internal/p2p/paygate/trust.go` and `internal/p2p/paygate/gate.go` use a post-pay threshold to decide friction.
   - That use is legitimate, but it should not collapse the broader meaning of reputation into one scalar payment rule.
   - This row therefore locks `payment trust` as one policy gate that reads reputation, not as the definition of reputation.

### Assessment

- `Reputation` should be kept.
- The correct action is `stabilize`:
  - keep agent/domain reputation earned from actual history,
  - keep owner-root trust separate as bootstrap context,
  - reserve full durable reputation redesign for `reputation v2`.

## Detailed Audit: Revocation & Trust Decay

### Audit Record

- Feature name: `Revocation & Trust Decay`
- Capability area: `External Collaboration & Economic Exchange`
- Product-path linkage: `Phase 1: Knowledge Exchange`, `Phase 2: Result Exchange with Controlled Execution`
- Current surface area: `docs/features/p2p-network.md`, `internal/p2p/handshake/security_events.go`, `internal/p2p/discovery/gossip.go`, `internal/p2p/reputation/*`
- Core value: `Reduce trust or revoke access quickly when runtime behavior becomes unsafe, without pretending every operational problem is already a durable reputation judgment.`
- Current problem: `Session invalidation and score updates are real, but the operator-facing model still needs a cleaner distinction between temporary operational safety actions and durable negative reputation.`
- Judgment: `stabilize`
- Execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas:
  - `Trust, Security & Policy`
- Secondary tracks:
  - `Stabilization Track`

### Findings

1. `Major` Immediate revocation is already a real runtime safety mechanism.
   - `docs/features/p2p-network.md` documents explicit invalidation reasons including `reputation_drop`, `repeated_failures`, `manual_revoke`, and `security_event`.
   - `internal/p2p/handshake/security_events.go` automatically invalidates sessions after repeated failures or reputation drops below threshold.

2. `Major` The current system already treats trust decay as an operational control path.
   - Repeated failures and security events can close access immediately.
   - This is appropriate for runtime safety and boundary control.
   - It should not be over-described as if the system already performs richer durable adjudication.

3. `Major` Durable negative reputation needs stricter semantics than operational decay.
   - `internal/p2p/reputation/store.go` records failures and timeouts into the score, which is enough for early exchange stabilization.
   - But the current surface does not yet justify treating every failure, timeout, or session revocation as a final durable reputational judgment across broader collaboration domains.

### Assessment

- `Revocation & Trust Decay` should be kept.
- The correct action is `stabilize`:
  - preserve immediate safety-oriented revocation and trust decay,
  - keep operational signals separate from durable negative reputation,
  - carry the harder adjudication model into later follow-on design work rather than overloading the current runtime.

## Assessment

All four rows remain `stabilize`: the capability family is real, but the operator-facing relationship model still needs consolidation.

## Follow-On Design Inputs

1. `reputation v2`
2. `pricing / negotiation / settlement` audit
3. `knowledge exchange runtime` end-to-end design
