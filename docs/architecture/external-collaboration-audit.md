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
| team formation / role coordination | Phase 3 | `docs/features/p2p-network.md`, `docs/features/multi-agent.md`, `internal/config/types_p2p.go`, `internal/config/types_orchestration.go`, `internal/cli/p2p/`, `internal/cli/agent/` | Detailed audit complete (`stabilize`) |
| workspace / shared artifacts | Phase 3-4 | `docs/features/p2p-network.md`, `docs/features/provenance.md`, `internal/config/types_p2p.go`, `internal/cli/p2p/`, `internal/cli/provenance/` | Detailed audit complete (`stabilize`) |

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

4. `Resolved` The post-pay trust threshold is now canonicalized at `0.8` across the payment-side runtime and operator-facing docs.
   - `paygate` and team payment now share the same canonical constant.
   - Exact-threshold behavior is inclusive on both sides of the payment path.
   - Operator-facing docs now describe `0.8` consistently and explicitly separate admission trust from payment trust.
   - References: `docs/features/p2p-network.md:506-525`, `docs/features/economy.md:20-23`, `docs/cli/p2p.md:540-543`, `internal/config/types_p2p.go:175-178`, `internal/p2p/trustpolicy/defaults.go:1-4`, `internal/p2p/paygate/trust.go:1-24`, `internal/p2p/paygate/gate.go:114-129`, `internal/p2p/team/payment.go:40-99`

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

4. `Resolved` Payment trust thresholds are now one stable payment-side model.
   - Admission trust and payment trust remain distinct by design, but payment-side post-pay routing now uses one canonical default of `0.8`.
   - The operator story now explicitly treats admission trust and payment trust as separate gates instead of competing defaults.
   - References: `docs/features/p2p-network.md:506-525`, `docs/features/economy.md:20-23`, `internal/config/types_p2p.go:175-178`, `internal/p2p/trustpolicy/defaults.go:1-4`, `internal/p2p/paygate/gate.go:114-129`, `internal/p2p/team/payment.go:40-99`

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
- The merged control-plane decision is explicit:
  - `p2p.pricing` is the provider-side public quote surface exposed to remote peers.
  - `economy.pricing` is the local dynamic pricing policy engine.
  - `economy.negotiation` is the local negotiation engine layered above the market path.
  - `economy.escrow` is the local escrow/policy engine for higher-friction settlement paths.
  - `paygate` and `settlement` remain the runtime payment path for P2P paid execution.
- The correct action is `merge`, with stabilization work following that convergence:
  - define one canonical operator story for quoting, negotiation, and settlement,
  - decide which surfaces belong to `P2P` and which belong to `Economy`,
  - reconcile static pricing vs dynamic pricing exposure,
  - reconcile settlement wording with the actual authorization-driven runtime.

## Detailed Audit: Team Formation / Role Coordination

### Audit Record

- Feature name: `team formation / role coordination`
- Capability area: `External Collaboration & Economic Exchange`
- Product-path linkage: `Phase 3: Leader-Led Team Execution`, `Phase 4: Long-Running Multi-Agent Projects`
- Current surface area: `docs/features/p2p-network.md`, `docs/features/multi-agent.md`, `docs/cli/p2p.md`, `internal/p2p/team/*`, `internal/app/bridge_team_*`, `internal/cli/p2p/team.go`, `internal/config/types_p2p.go`
- Core value: `Form leader-led external teams, assign roles, coordinate delegated work, and connect reputation, budget, escrow, and health into one collaborative execution model.`
- Current problem: `The subsystem is real and broad in code, but operator-facing control is largely placeholder, and several documented semantics do not match the actual coordinator, conflict, and payment logic.`
- Judgment: `stabilize`
- Execution track: `Leader-Led Team Execution Track`
- Secondary capability areas:
  - `Execution, Continuity & Accountability`
  - `Trust, Security & Policy`
- Secondary tracks:
  - `Stabilization Track`

### Findings

1. `Major` Team coordination is a real subsystem worth keeping.
   - The codebase includes a real `Coordinator`, `Team` model, role system, assignment strategies, conflict strategies, health monitor, payment negotiation, and bridges into budget, escrow, reputation, shutdown, metrics, and workspace flows.
   - This is not a stubbed idea; it is an actual Phase 3 subsystem with meaningful runtime integration.
   - References: `internal/p2p/team/coordinator.go:18-212`, `internal/p2p/team/team.go:17-189`, `internal/app/bridge_team_budget.go`, `internal/app/bridge_team_escrow.go`, `internal/app/bridge_team_reputation.go`, `internal/app/bridge_team_shutdown.go`

2. `Major` The documented/live operator surface is not aligned with the implementation.
   - `lango p2p team list/status/disband` are documented as active management commands with concrete examples.
   - The actual CLI subcommands are runtime-only placeholders that return empty or “team not found” results and tell the operator to use server API endpoints.
   - Those referenced `/api/p2p/teams/<id>` endpoints are not registered in the current P2P HTTP routes.
   - References: `docs/cli/p2p.md:463-532`, `internal/cli/p2p/team.go:15-132`, `internal/app/p2p_routes.go:28-37`

3. `Major` Conflict-resolution docs overstate what the code currently does.
   - The docs describe `trust_weighted` as selecting the highest-trust result.
   - The implementation actually chooses the shortest-duration successful result as a trust proxy.
   - The docs describe `majority_vote` as selecting the most common result.
   - The implementation currently returns the first successful result.
   - References: `docs/features/p2p-network.md:618-632`, `internal/p2p/team/conflict.go:18-57`

4. `Resolved` Team payment coordination now uses the same inclusive `0.8` post-pay threshold as the broader external payment path.
   - Team payment mode switching is now aligned with the shared payment-side threshold rather than carrying its own competing default.
   - References: `docs/features/p2p-network.md:711-719`, `docs/features/multi-agent.md:388-392`, `docs/cli/p2p.md:540-543`, `internal/p2p/team/payment.go:40-99`, `internal/p2p/trustpolicy/defaults.go:1-4`

5. `Major` Team formation exists as an internal coordinator flow, but not as a stable user-facing formation path.
   - `FormTeamRequest` and `Coordinator.FormTeam()` provide a concrete internal formation API that selects workers from the pool by capability.
   - The public CLI/API surface does not yet expose a matching live formation/control path.
   - This makes the feature real for runtime integrations but not yet stable as an operator-facing product surface.
   - References: `internal/p2p/team/coordinator.go:124-212`, `internal/cli/p2p/team.go:15-132`, `internal/app/p2p_routes.go:28-37`

### Assessment

- `Team formation / role coordination` should be kept. It is central to the Phase 3 and Phase 4 product path.
- The correct action is `stabilize`, not `defer`:
  - reconcile CLI/docs with the actual live control surface,
  - either add the documented team HTTP endpoints or stop documenting them,
  - reconcile conflict-strategy descriptions with the current implementation,
  - define one stable operator path for forming and inspecting live teams.

## Detailed Audit: Workspace / Shared Artifacts

### Audit Record

- Feature name: `workspace / shared artifacts`
- Capability area: `External Collaboration & Economic Exchange`
- Product-path linkage: `Phase 3: Leader-Led Team Execution`, `Phase 4: Long-Running Multi-Agent Projects`
- Current surface area: `docs/features/p2p-network.md`, `docs/cli/p2p.md`, `internal/p2p/workspace/*`, `internal/p2p/gitbundle/*`, `internal/p2p/provenanceproto/*`, `internal/cli/p2p/workspace.go`, `internal/cli/p2p/git.go`, `internal/cli/p2p/provenance.go`, `internal/app/wiring_workspace.go`, `internal/app/bridge_workspace_team.go`, `internal/app/tools_workspace.go`
- Core value: `Provide collaborative external workspaces, shared artifact exchange, and auditable handoff mechanisms for code and provenance.`
- Current problem: `The runtime subsystems are real, but the operator surface is uneven: provenance exchange is live, while workspace and git bundle controls are mostly documented as direct commands even though the runtime expects server-backed or tool-backed flows.`
- Judgment: `stabilize`
- Execution track: `Leader-Led Team Execution Track`
- Secondary capability areas:
  - `Execution, Continuity & Accountability`
  - `Trust, Security & Policy`
- Secondary tracks:
  - `Stabilization Track`

### Findings

1. `Major` Workspace and shared-artifact infrastructure is real and broad enough to keep.
   - The codebase includes a persistent workspace manager, workspace gossip, contribution tracking, chronicler hooks, git bundle services, provenance exchange protocol handlers, and team-to-workspace bridges.
   - This is a genuine Phase 3/4 capability, not a speculative placeholder.
   - References: `internal/p2p/workspace/manager.go:17-223`, `internal/p2p/workspace/gossip.go:14-160`, `internal/p2p/gitbundle/bundle.go:51-470`, `internal/app/bridge_workspace_team.go:14-102`, `internal/app/p2p_routes.go:36-37`

2. `Major` Workspace CLI docs overstate direct command behavior.
   - The docs show `workspace create/list/status/join/leave` as if they return live workspace data.
   - The actual CLI commands are runtime placeholders that direct the operator to `lango serve` and agent tools instead of performing the documented live action.
   - References: `docs/cli/p2p.md:601-732`, `internal/cli/p2p/workspace.go:15-204`

3. `Major` Git bundle CLI docs also overstate direct command behavior.
   - The docs present `git init/log/diff/push/fetch` as if they directly operate on live workspace repos.
   - The actual CLI commands mostly print “requires a running server” and defer to agent tools or server-backed flows.
   - The underlying gitbundle service is real; the operator-facing command story is what drifts.
   - References: `docs/cli/p2p.md:738-888`, `internal/cli/p2p/git.go:13-151`, `internal/app/tools_workspace.go`

4. `Major` Provenance bundle exchange is the one genuinely live operator surface in this family, which makes the shared-artifact model uneven.
   - `lango p2p provenance push/fetch` actually talks to gateway-backed `/api/p2p/provenance/*` endpoints.
   - That means one shared-artifact path is live and concrete, while neighboring workspace/git surfaces remain mostly placeholder or tool-backed.
   - References: `docs/cli/p2p.md:327-349`, `internal/cli/p2p/provenance.go:15-142`, `internal/app/p2p_routes.go:36-37`, `internal/app/p2p_routes.go:183-260`

5. `Major` The chronicler path is documented as persistent graph-triple capture, but the current app wiring still leaves the triple adder pending.
   - The docs describe chronicler persistence as an available workspace feature.
   - The workspace wiring currently instantiates the chronicler with a `nil` triple adder and logs that the triple adder is still pending.
   - So the concept is present, but the operator-facing statement is ahead of the fully wired default runtime path.
   - References: `docs/features/p2p-network.md:780-789`, `internal/app/wiring_workspace.go:102-109`, `internal/p2p/workspace/chronicler.go:24-46`

6. `Major` Shared artifacts are implemented through multiple overlapping mechanisms without one canonical operator story.
   - Workspaces manage runtime membership and message flow.
   - Git bundles manage code-state transfer.
   - Provenance bundles manage signed historical/session exchange.
   - Team bridges auto-create workspaces and record contributions.
   - All of these pieces are individually reasonable, but the operator-facing path for “how external agents actually share artifacts” is not yet unified.
   - References: `internal/p2p/workspace/manager.go`, `internal/p2p/gitbundle/bundle.go`, `internal/p2p/provenanceproto/protocol.go`, `internal/app/bridge_workspace_team.go`

### Assessment

- `Workspace / shared artifacts` should be kept. It is part of the Phase 3 and Phase 4 collaboration story.
- The correct action is `stabilize`, not `merge` or `defer`:
  - reconcile workspace/git CLI docs with the actual live control path,
  - decide which artifact operations are server-backed, which are agent-tool-backed, and which are true direct CLI commands,
  - clarify provenance vs workspace vs git-bundle responsibilities,
  - either fully wire chronicler persistence or narrow the user-facing claim,
  - publish one canonical operator story for shared artifact exchange.

## Next Plan

The next implementation plan after this document lands should perform the detailed audit for the first row:

- convert the completed external-collaboration audit into a prioritized stabilization and consolidation plan
