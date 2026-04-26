# Identity Trust Reputation Audit

## Purpose

This document is the next detailed audit ledger under the Lango master document.

It now locks the landed Reputation V2 model for the identity, trust, and reputation boundary that `knowledge exchange v1` already depends on:

- identity continuity,
- trust entry,
- reputation,
- revocation and trust decay.

Its purpose is to keep the operator-facing relationship model aligned with the landed runtime contract so later pricing, runtime, and settlement work inherits one consistent interpretation.

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

The landed V2 contract now makes that relationship explicit in runtime terms:

- `trustScore` remains the composite compatibility score used for broad runtime comparisons.
- `earnedTrustScore` is derived from actual collaboration history only and excludes temporary operational incidents.
- `durableNegativeUnits` tracks lasting negative outcomes. Standard failures add `1`, while adjudicated failures add `2`.
- `temporarySafetySignals` tracks operational incidents such as timeouts or unhealthy-member events without automatically turning them into durable reputation damage.
- canonical trust-entry states are `bootstrap`, `established`, `review`, and `temporarily_unsafe`.
- the current runtime policy consumes `BootstrapTrustScore`, `MinEarnedTrustScore`, and `MaxTemporarySafetySignals`; `OwnerRootTrusted` remains available in the contract for future callers but is not enabled by the default runtime wiring today.

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
- Current problem: `Gateway auth, handshake approval, firewall admission, and payment friction are all real, so the remaining work is keeping the docs and operator story aligned with one canonical trust-entry model instead of letting each consumer drift.`
- Judgment: `stabilize`
- Execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas:
  - `Trust, Security & Policy`
- Secondary tracks:
  - `Stabilization Track`

### Findings

1. `Major` Admission trust now has one canonical runtime contract.
   - `internal/gateway/auth.go` establishes gateway session-based auth when OIDC is configured.
   - `internal/p2p/reputation/contract.go` evaluates trust entry into `bootstrap`, `established`, `review`, and `temporarily_unsafe`.
   - `docs/features/p2p-network.md` describes the P2P approval pipeline as firewall ACL, reputation check, and owner approval.
   - `internal/app/p2p_routes.go` protects `/api/p2p` with gateway auth whenever auth is configured.

2. `Major` Admission trust and payment trust stay separate, but they now read the same canonical trust entry.
   - Admission trust answers whether the peer is allowed in at all.
   - Payment trust answers whether a paid invocation can use post-pay or must use upfront payment.
   - `internal/app/runtime_reputation.go` returns the bootstrap floor for new peers, the earned trust score for returning peers, and zero post-pay trust unless the peer is `established`.
   - `internal/p2p/paygate/gate.go` therefore enforces payment-tier routing after the invocation path already exists, which is a different question from handshake or firewall entry.

3. `Major` New peers and returning peers now separate cleanly in the runtime contract.
   - `internal/app/runtime_reputation.go` treats first-time peers as `bootstrap`.
   - `internal/app/runtime_reputation.go` only auto-approves known peers when they are returning, `established`, allowed, and do not require review.
   - `internal/p2p/firewall/firewall.go` blocks `review` and `temporarily_unsafe` returning peers before ACL allow rules can admit them.

### Assessment

- `Trust Entry` should be kept.
- The correct action is `stabilize`:
  - document gateway auth, firewall entry, and handshake/session controls as one admission model,
  - preserve the separation between admission trust and payment trust,
  - keep new peers bootstrap-trusted while reserving lower-friction paths for earned `established` trust.

## Detailed Audit: Reputation

### Audit Record

- Feature name: `Reputation`
- Capability area: `External Collaboration & Economic Exchange`
- Product-path linkage: `Phase 1: Knowledge Exchange`, `Phase 2: Result Exchange with Controlled Execution`
- Current surface area: `docs/features/p2p-network.md`, `docs/cli/p2p.md`, `internal/p2p/reputation/*`, `internal/app/p2p_routes.go`, `internal/p2p/team/payment.go`
- Core value: `Capture exchange history so trust can be earned from actual outcomes instead of being inherited wholesale from owner identity or one-time entry approval.`
- Current problem: `The canonical V2 model now exists in code, so the remaining risk is documentation drift rather than missing runtime semantics.`
- Judgment: `stabilize`
- Execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas:
  - `Trust, Security & Policy`
- Secondary tracks:
  - `Stabilization Track`

### Findings

1. `Major` Reputation now exposes separated runtime signals rather than only one scalar score.
   - `internal/p2p/reputation/store.go` persists successful exchanges, durable negative units, temporary safety signals, trust score, first-seen, and last-interaction state.
   - `internal/p2p/reputation/contract.go` derives `earnedTrustScore`, `durableNegativeUnits`, and `temporarySafetySignals` into `CanonicalSignals`.
   - `internal/app/p2p_routes.go` exposes read-only reputation details through `/api/p2p/reputation`.
   - `docs/gateway/http-api.md` and `docs/cli/p2p.md` present reputation as a first-class operator-visible concept.

2. `Major` Reputation should remain agent/domain-earned rather than owner-inherited.
   - The current track language already distinguishes owner-root trust from agent-specific reputation.
   - The reputation store is keyed by `peerDID` and updated from actual exchange outcomes, which aligns with an earned-reputation model rather than a root-trust inheritance model.
   - The current runtime wiring does not set `OwnerRootTrusted`, so bootstrap and earned trust are the active policy inputs today.

3. `Major` Durable and temporary negatives now have stronger, different semantics.
   - `internal/p2p/reputation/contract.go` records adjudicated failures with a stronger durable penalty than standard failures.
   - `internal/p2p/reputation/contract.go` records operational incidents separately and keeps them out of `earnedTrustScore`.
   - This keeps `payment trust` and other consumers reading reputation as policy inputs without redefining reputation itself.

### Assessment

- `Reputation` should be kept.
- The correct action is `stabilize`:
  - keep agent/domain reputation earned from actual history,
  - keep owner-root trust separate as bootstrap context,
  - keep the landed V2 contract as the canonical baseline for later operator and dispute-surface work.

## Detailed Audit: Revocation & Trust Decay

### Audit Record

- Feature name: `Revocation & Trust Decay`
- Capability area: `External Collaboration & Economic Exchange`
- Product-path linkage: `Phase 1: Knowledge Exchange`, `Phase 2: Result Exchange with Controlled Execution`
- Current surface area: `docs/features/p2p-network.md`, `internal/p2p/handshake/security_events.go`, `internal/p2p/discovery/gossip.go`, `internal/p2p/reputation/*`
- Core value: `Reduce trust or revoke access quickly when runtime behavior becomes unsafe, without pretending every operational problem is already a durable reputation judgment.`
- Current problem: `Session invalidation and score updates are real, and the runtime now distinguishes temporary operational safety from durable negative reputation, so the remaining work is broader adoption rather than inventing new semantics.`
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

2. `Major` The current system now treats trust decay as an explicitly operational control path.
   - `internal/p2p/reputation/contract.go` exposes `temporarySafetySignals` separately from durable negatives.
   - `internal/p2p/firewall/firewall.go` treats `temporarily_unsafe` as an immediate runtime block.
   - `internal/app/bridge_team_reputation.go` records operational incidents from unhealthy members and kicks only when the refreshed trust entry no longer allows runtime collaboration.

3. `Major` Durable negative reputation now uses stricter semantics than operational decay.
   - `internal/p2p/reputation/contract.go` distinguishes standard failures from adjudicated failures.
   - `internal/p2p/reputation/store.go` still keeps the composite score for compatibility, but V2 readers can now avoid treating every timeout or unhealthy event as durable damage.
   - That is the right contract for early runtime safety without over-claiming broader adjudication coverage.

### Assessment

- `Revocation & Trust Decay` should be kept.
- The correct action is `stabilize`:
  - preserve immediate safety-oriented revocation and trust decay,
  - keep operational signals separate from durable negative reputation,
  - carry the harder adjudication model into later follow-on design work rather than overloading the current runtime.

## Assessment

All four rows remain `stabilize`: the capability family is real, but the operator-facing relationship model still needs consolidation.

## Follow-On Design Inputs

1. broader owner-root-aware policy adoption on top of the landed V2 contract
2. richer dispute-to-reputation integration beyond the current adjudicated failure hook
3. broader operator-facing trust, review, and recovery surfaces
