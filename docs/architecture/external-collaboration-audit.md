# External Collaboration & Economic Exchange Audit

## Purpose

This document is the first detailed audit ledger under the Lango master document.

It exists to review the product area that most directly defines Lango:

- P2P identity,
- trust,
- reputation,
- pricing,
- negotiation,
- settlement,
- team formation,
- shared artifacts.

## Relationship to the Master Document

This audit sits underneath `docs/architecture/master-document.md` and must use that document's constitution, capability taxonomy, audit vocabulary, and track-routing rules.

It does not redefine what Lango is, replace the product constitution, or create new top-level capability areas or execution tracks. Its role is to apply the master document's framework to the external collaboration capability area in detailed ledger form.

## Document Ownership

- Primary capability area: `External Collaboration & Economic Exchange`
- Primary execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas: `Execution, Continuity & Accountability`
- Secondary tracks: `Leader-Led Team Execution Track`
- Rows may override to `Leader-Led Team Execution Track` when the main responsibility is team formation, role coordination, delegated budget control, or shared artifacts for Phase 3 execution work and current Phase 4 collaboration or execution work.
- For provenance, ledgers, workflow continuation, or accountability-heavy rows, detailed follow-on audit work may classify `Execution, Continuity & Accountability` as the primary capability area using the master document's Phase 4 tie-break.

## Audit Order

1. P2P identity / trust / reputation
2. pricing / negotiation / settlement
3. team formation / role coordination
4. workspace / shared artifacts

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

Each feature family should be judged by:

- capability area fit,
- product-path fit,
- current user-facing surface,
- duplication risk,
- trust or policy gaps,
- judgment,
- owning track.

Allowed judgments:

- `keep`
- `stabilize`
- `merge`
- `defer`
- `remove`

## Current Surface Map

| Feature family | Primary phase | Current surface clues | Audit status |
| --- | --- | --- | --- |
| P2P identity / trust / reputation | Phase 1 | `docs/features/p2p-network.md`, `docs/features/economy.md`, `internal/config/types_p2p.go`, `internal/cli/p2p/`, `internal/cli/settings/forms_p2p.go` | Detailed audit complete (`stabilize`) |
| pricing / negotiation / settlement | Phase 1-2 | `docs/features/economy.md`, `docs/payments/usdc.md`, `docs/payments/x402.md`, `internal/config/types_economy.go`, `internal/cli/economy/`, `internal/cli/payment/` | Detailed audit complete (`merge`) |
| team formation / role coordination | Phase 3 | `docs/features/p2p-network.md`, `docs/features/multi-agent.md`, `internal/config/types_p2p.go`, `internal/config/types_orchestration.go`, `internal/cli/p2p/`, `internal/cli/agent/` | Ready for detailed audit |
| workspace / shared artifacts | Phase 3-4 | `docs/features/p2p-network.md`, `docs/features/provenance.md`, `internal/config/types_p2p.go`, `internal/cli/p2p/`, `internal/cli/provenance/` | Ready for detailed audit |

## Baseline Decisions Already Locked

- External collaboration is economically native.
- Trust is mixed: cryptographic continuity first, transaction history at the center.
- Root accountability belongs to the owner.
- Agent-level reputation stays separate from owner-level root trust.
- Early trade is bounded by allowlists plus explicit exportability policy.
- The default early external exchange is deliverable-oriented, not broad execution.
- On-chain stablecoin is the trust anchor for settlement.
- Off-chain accrual opens only after trust is earned.
- Team formation is leader-led by default.
- Shared artifacts are leader-owned and selectively exposed by scope.

## Detailed Audit: P2P Identity / Trust / Reputation

### Audit Record

- Feature name: `P2P identity / trust / reputation`
- Capability area: `External Collaboration & Economic Exchange`
- Product-path linkage: `Phase 1: Knowledge Exchange`, `Phase 2: Result Exchange with Controlled Execution`
- Current surface area: `docs/features/p2p-network.md`, `docs/features/economy.md`, `internal/p2p/identity/*`, `internal/p2p/handshake/*`, `internal/p2p/reputation/store.go`, `internal/p2p/paygate/*`, `internal/cli/p2p/identity.go`, `internal/cli/p2p/reputation.go`, `internal/app/p2p_routes.go`, `internal/config/types_p2p.go`
- Core value: `Establish cryptographic continuity, gate remote access by trust, and route payment friction by peer reputation.`
- Current problem: `The runtime is real and cross-wired, but the operator-facing model drifts across CLI, docs, API auth semantics, DID versions, and payment trust defaults.`
- Judgment: `stabilize`
- Execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas:
  - `Trust, Security & Policy`
  - `Execution, Continuity & Accountability`
- Secondary tracks:
  - `Stabilization Track`

### Findings

1. `Major` CLI identity surface does not actually expose DID even though docs and command help say it does.
   - The published docs describe `lango p2p identity` as a DID identity surface, and the CLI help repeats that claim.
   - The current implementation prints only peer ID, key storage, and listen addresses in text mode and JSON mode.
   - References: `docs/features/p2p-network.md:398-410`, `docs/features/p2p-network.md:433-435`, `internal/cli/p2p/identity.go:17-19`, `internal/cli/p2p/identity.go:41-57`, `internal/app/p2p_routes.go:170-178`

2. `Major` REST auth semantics for identity and reputation are documented as public, but the runtime protects the whole `/api/p2p` subtree when auth is configured.
   - The docs say the read-only P2P endpoints are public and unauthenticated.
   - The route registration applies `gateway.RequireAuth(auth)` to `/api/p2p`, and the code comment only treats the endpoints as public when auth is `nil` in dev mode.
   - References: `docs/features/p2p-network.md:392-420`, `internal/app/p2p_routes.go:25-35`

3. `Major` The runtime supports two DID modes, but the published identity story still documents only the legacy v1 DID.
   - The user-facing P2P docs present only `did:lango:<hex-compressed-pubkey>`.
   - The codebase supports `did:lango:v2:<hash>`, identity bundles, bundle resolution, DID aliasing, and runtime selection of the bundle provider when identity material is available.
   - References: `docs/features/p2p-network.md:46-54`, `internal/p2p/identity/identity.go:20-27`, `internal/app/wiring_p2p.go:186-204`, `internal/app/wiring_p2p.go:261-279`

4. `Major` The post-pay trust threshold has a documented default of `0.8`, but the runtime falls back to `0.7` when the config leaves the field unset.
   - The docs and config comments present `postPayMinScore` as `0.8`.
   - The runtime builds `trustCfg` from `paygate.DefaultTrustConfig()` and only overrides it if the config value is greater than zero.
   - `paygate.DefaultTrustConfig()` resolves to `0.7`.
   - References: `docs/features/p2p-network.md:506-525`, `internal/config/types_p2p.go:175-178`, `internal/p2p/paygate/trust.go:6-23`, `internal/app/wiring_p2p.go:465-469`

5. `Major` Trust is operationally wired end-to-end, but the operator-facing model is fragmented across admission, session invalidation, and payment-tier routing.
   - Firewall admission rejects known low-score peers but still lets new peers with score `0` through.
   - Reputation changes can revoke sessions through the security event handler.
   - Payment friction uses a separate trust threshold to switch between prepay and postpay.
   - The runtime behavior is coherent enough to keep, but not coherent enough to present as one stable operator model.
   - References: `internal/p2p/firewall/firewall.go:158-167`, `internal/p2p/reputation/store.go:101-129`, `internal/app/wiring_p2p.go:422-433`, `internal/p2p/paygate/gate.go:112-127`

### Assessment

- `Identity` is a real core capability worth keeping. DID-to-peer binding, handshake versioning, session issuance, and reputation persistence are already wired.
- `Trust / reputation` should not be removed or merged away; it is part of the Phase 1 and Phase 2 product path.
- The correct action is `stabilize`, not `defer`:
  - reconcile the public identity surface,
  - reconcile auth semantics for `/api/p2p/*`,
  - reconcile v1/v2 DID documentation,
  - reconcile `postPayMinScore` default drift,
  - publish one canonical operator-facing trust model.

## Detailed Audit: Pricing / Negotiation / Settlement

### Audit Record

- Feature name: `pricing / negotiation / settlement`
- Capability area: `External Collaboration & Economic Exchange`
- Product-path linkage: `Phase 1: Knowledge Exchange`, `Phase 2: Result Exchange with Controlled Execution`
- Current surface area: `docs/features/p2p-network.md`, `docs/features/economy.md`, `docs/payments/usdc.md`, `docs/payments/x402.md`, `internal/p2p/paygate/*`, `internal/p2p/settlement/*`, `internal/economy/pricing/*`, `internal/economy/negotiation/*`, `internal/cli/p2p/pricing.go`, `internal/cli/economy/*`, `internal/app/wiring_p2p.go`, `internal/app/wiring_economy.go`
- Core value: `Turn peer trust and tool value into a payable external exchange path, including quoting, negotiation, and on-chain settlement.`
- Current problem: `The capability exists, but it is split across a P2P payment-gate path and a separate economy subsystem, so pricing, negotiation, and settlement do not present one canonical operator model.`
- Judgment: `merge`
- Execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas:
  - `Trust, Security & Policy`
  - `Execution, Continuity & Accountability`
- Secondary tracks:
  - `Consolidation Track`
  - `Stabilization Track`

### Findings

1. `Major` The operator-facing pricing/settlement surface is split across `p2p` and `economy`, not expressed as one coherent external-exchange model.
   - The P2P surface exposes static tool pricing through `p2p.pricing.perQuery` and `p2p.pricing.toolPrices`, plus a payment gate and settlement service.
   - The economy surface separately exposes dynamic pricing, negotiation, escrow, and on-chain escrow tooling.
   - This makes the same conceptual capability appear under two control planes with different defaults and narratives.
   - References: `docs/features/p2p-network.md:464-571`, `docs/features/economy.md:11-21`, `internal/cli/p2p/pricing.go:21-76`, `internal/cli/economy/pricing.go:11-38`, `internal/cli/economy/negotiate.go:11-40`

2. `Major` Negotiation is implemented, but it is under-surfaced for the P2P knowledge-exchange path.
   - The economy negotiation engine is real and wired into the P2P protocol handler.
   - The exposed operator surface is still mostly economy-config and economy-tool oriented, rather than a clear P2P-facing transaction path.
   - In other words, the runtime can negotiate, but the track-level surface does not yet read like one canonical external market flow.
   - References: `internal/app/wiring_economy.go:120-180`, `internal/app/wiring_economy.go:331-383`, `internal/economy/tools.go:209-260`, `docs/features/economy.md:112-154`

3. `Major` Settlement documentation overstates or blurs how the current runtime actually authorizes payment.
   - The P2P docs say wallet addresses are derived from peer DIDs for settlement.
   - The payment-gate path actually validates an explicit `paymentAuth`, checks that `auth.To` matches the local wallet address, then hands the authorization to the settlement service.
   - The economic path is real, but the user-facing story is not phrased in the same terms as the runtime.
   - References: `docs/features/p2p-network.md:487-571`, `internal/p2p/paygate/gate.go:112-185`, `internal/p2p/settlement/service.go:85-124`

4. `Major` Payment trust thresholds are still not one stable model.
   - Admission trust and payment trust are distinct by design, but the defaults and operator story are not reconciled.
   - `minTrustScore` gates admission around `0.3`, while post-pay routing falls back to `0.7` in code even though the docs and config comments still present `0.8`.
   - The result is a real but fragmented trust-to-friction model.
   - References: `docs/features/p2p-network.md:506-525`, `internal/config/types_p2p.go:175-178`, `internal/p2p/paygate/trust.go:6-23`, `internal/app/wiring_p2p.go:465-469`

5. `Major` The current P2P pricing API exposes only static quotes, while dynamic pricing and negotiation live elsewhere.
   - `/api/p2p/pricing` and `lango p2p pricing` surface only `perQuery` and `toolPrices`.
   - The dynamic pricing engine can compute trust-sensitive quotes, but that logic is not what the P2P surface exposes today.
   - This reinforces that the current capability is not missing, but split.
   - References: `internal/app/p2p_routes.go:142-167`, `internal/cli/p2p/pricing.go:21-76`, `internal/economy/pricing/engine.go:18-141`

### Assessment

- `Pricing / negotiation / settlement` is a core Phase 1 and Phase 2 capability and should be kept.
- The primary problem is not absence. The problem is duplicated or fragmented control planes:
  - `p2p.pricing` and `paygate/settlement`
  - `economy.pricing`
  - `economy.negotiation`
  - `economy.escrow`
- The correct action is `merge`, with stabilization work following that convergence:
  - define one canonical operator story for quoting, negotiation, and settlement,
  - decide which surfaces belong to `P2P` and which belong to `Economy`,
  - reconcile static pricing vs dynamic pricing exposure,
  - reconcile settlement wording with the actual authorization-driven runtime,
  - reconcile trust thresholds and payment-tier defaults.

## Next Plan

The next implementation plan after this document lands should perform the detailed audit for the first row:

- team formation / role coordination
