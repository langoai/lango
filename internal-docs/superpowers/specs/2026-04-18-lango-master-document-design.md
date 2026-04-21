# Lango Master Document Design

Date: 2026-04-18
Status: Proposed and validated through brainstorming
Scope: Product constitution, audit framework, and track portfolio for the top-level Lango master document

## Purpose

This document defines the structure and governing principles for a top-level Lango master document.

The master document is not a feature list and not a stabilization-only plan. It exists to:

- define what Lango fundamentally is,
- define how current features are judged,
- define which work belongs in which execution track, and
- keep future implementation plans aligned with a single product story.

The existing stabilization plan remains useful, but it should sit underneath this master document as one execution track rather than acting as the top-level strategy.

## Problem Statement

Lango currently exposes a broad set of features across runtime UX, knowledge systems, multi-agent orchestration, P2P networking, security, economy, on-chain settlement, provenance, and automation. That breadth creates two risks:

1. the product story becomes unclear,
2. the codebase, docs, settings surface, and CLI/TUI surface can grow without a stable decision framework.

The master document must solve both risks by combining product constitution, feature audit rules, and execution-track routing in one place.

## Product Constitution

The master document should lock the following constitutional principles:

1. Lango is a sovereign peer-to-peer agent network, not just a local agent runtime.
2. External collaboration between sovereign agents is the central product identity.
3. Internal collaboration between agents owned by the same person, company, or institution matters, but it is a reduced form of the external collaboration model.
4. External collaboration is economically native. Computation, expertise, time, and risk are not free.
5. On-chain settlement is not the product itself. It is one trusted settlement mechanism inside a larger economic coordination system.
6. The default external collaboration structure is leader-led. A leader agent discovers peers, evaluates trust, forms teams, negotiates work, controls budget, and coordinates settlement.
7. The long-term product target is high autonomy, but the practical near-term model is delegated autonomy.
8. User private conversations, confidential material, and sensitive internal information are never tradeable assets.
9. Tradeable knowledge and deliverables must stay inside an allowlist plus explicit exportability policy.
10. The first economic activity is knowledge exchange and result exchange, not broad remote execution.
11. Trust starts from cryptographic continuity of identity, but reputation should be centered on real collaboration and transaction history.
12. The owner carries root identity and final accountability, while individual agents accumulate separate role-specific and domain-specific reputation.

## Primary Product Path

The master document should define a staged product path:

### Phase 1: Knowledge Exchange

- Sovereign external agents participate through pseudonymous but cryptographically continuous identities.
- The first market activity is expertise access plus reviewable deliverables.
- The default experience is delivery of outputs such as summaries, research notes, design drafts, and code drafts.
- Tradeable artifacts remain bounded by allowlists and exportability policy.
- Default settlement is small upfront payment plus approval-based final settlement, anchored by on-chain stablecoin.

### Phase 2: Result Exchange with Controlled Execution

- Repeat interactions and stronger trust allow more structured result exchange.
- Limited execution may be opened only for higher-trust relationships with explicit approval and strong policy controls.
- Off-chain accrual, dynamic credit limits, clearer acceptance criteria, and formalized dispute handling mature here.

### Phase 3: Leader-Led Team Execution

- A leader agent forms an external team around a shared goal.
- Team composition uses standard role templates plus optional custom roles.
- Budgeting, contracting, and settlement remain leader-controlled by default, with narrow delegated authority when needed.
- Workspace sharing is selective and leader-owned by default.
- Artifact exchange is snapshot/package oriented by default, not unrestricted live co-editing.

### Phase 4: Long-Running Multi-Agent Projects

- Collaboration expands from one-off tasks to long-running, multi-session, multi-milestone projects.
- Provenance, ledgers, recurring settlement, durable work records, and persistent shared artifacts become central.
- This is where Lango begins to resemble an economically native multi-agent organization layer.

## Capability Areas

The master document should group features by capability area rather than by config file or package tree.

### 1. Sovereign Runtime & Operator UX

Includes providers, agent runtime, channels, cockpit, settings, status, session handling, logging, and core tool surfaces.

This area is about operating Lango clearly and confidently.

### 2. Knowledge & Intelligence Capital

Includes knowledge, observational memory, embedding/RAG, graph, ontology, librarian, agent memory, skills, and context systems.

This area is about how Lango learns, remembers, generalizes, and forms tradeable intelligence capital.

### 3. External Collaboration & Economic Exchange

Includes P2P identity, trust, reputation, pricing, negotiation, settlement, team formation, and shared artifact exchange.

This is the primary differentiator of the product and should be the first audited capability area.

### 4. Trust, Security & Policy

Includes auth, signer/KMS, privacy boundaries, exportability policy, approvals, gatekeeper, exec safety, sandboxing, and auditability.

This area defines what can be trusted, what can be exported, and what must be blocked.

### 5. Execution, Continuity & Accountability

Includes workflow, background tasks, cron, RunLedger, provenance, and hooks.

This area supports delegated work, continuation, traceability, and accountable execution.

### 6. Extensibility & Developer Clarity

Includes MCP, extension packs, architecture boundaries, docs truthfulness, CLI surface clarity, and duplicate cleanup.

This area keeps the system extensible and understandable for developers and operators.

## Feature Audit Framework

The master document should judge features by capability value, not by existence.

### Evaluation Principles

- A feature is valuable if it directly strengthens the core product story.
- The strongest test is whether it materially supports:
  - sovereign external collaboration,
  - knowledge exchange,
  - result exchange,
  - leader-led team execution, or
  - long-running multi-agent collaboration.
- Features with overlapping responsibility should be considered merge candidates.
- Features aligned with the vision but weak in reliability, defaults, docs, trust boundaries, or UX should be stabilization candidates.
- Features that fit the vision but are too early for the current product path should be deferred.
- Features that add complexity without strengthening the core story should be removal candidates.

### Allowed Audit Judgments

- `keep`
- `stabilize`
- `merge`
- `defer`
- `remove`

### Audit Row Structure

Each audited feature or feature family should include:

- feature name,
- capability area,
- product-path linkage,
- current surface area,
- core value,
- current problem,
- judgment,
- execution track.

## Track Portfolio

The master document should route concrete work into a small number of execution tracks.

### Stabilization Track

Truth alignment, durability, defaults, hardening, observability, and production correctness.

The current stabilization planning work belongs here.

### Consolidation Track

Merges duplicate features, duplicate config surfaces, duplicate commands, and overlapping responsibilities.

This track directly addresses the feeling that the product has too many loosely connected surfaces.

### UX Clarity Track

Simplifies settings, cockpit, CLI, docs, and defaults to create a much clearer operator experience without requiring a ground-up rewrite.

### P2P Knowledge Exchange Track

Defines the first concrete market activity for external sovereign agents.

This track should cover:

- identity,
- trust,
- reputation,
- exportability,
- pricing,
- negotiation,
- settlement,
- dispute handling,
- reviewable deliverables.

### Leader-Led Team Execution Track

Builds the next layer above knowledge exchange:

- team formation,
- role coordination,
- scoped budget delegation,
- shared artifacts,
- milestone-based execution.

### Developer Clarity Track

Improves architecture boundaries, internal naming, documentation truthfulness, and the ability to understand what Lango is without reading the whole codebase.

## External Collaboration Audit Order

The first capability area audit should be `External Collaboration & Economic Exchange`.

Its internal audit order should be:

1. P2P identity / trust / reputation
2. pricing / negotiation / settlement
3. team formation / role coordination
4. workspace / shared artifacts

## External Collaboration Decisions Captured So Far

The master document design should preserve these decisions as baseline assumptions for later track work.

### Identity, Trust, and Reputation

- Trust is mixed:
  - cryptographic identity continuity is the base,
  - real transaction and collaboration history is the center of reputation,
  - endorsements and guarantees are secondary signals.
- Reputation should be mixed:
  - a high-level overall reputation may be shown,
  - real decision-making should rely on domain-specific reputation.
- Positive signals can be reflected automatically.
- Negative signals should require dispute resolution or explicit adjudication before durable reputation impact.
- Dispute handling should be leader-first, with human escalation for high-risk, high-value, or contested cases.
- The dispute environment should still be rule-bound and difficult to manipulate through signed logs, provenance, acceptance criteria, escrow state, and immutable receipts.
- Root accountability belongs to the owner.
- Agent-level role and domain reputation should remain separate from owner-level root trust.
- New agents should begin under constrained terms:
  - small prepaid or escrowed activity first,
  - broader participation only after trust, guarantees, collateral, or extra verification.
- New agents under an already trusted owner should receive limited trust inheritance, not full trust inheritance.

### Knowledge Exchange

- `v1` trade should be bounded by allowlists plus explicit exportability policy.
- The user defines policy and boundary conditions.
- The leader agent may classify assets as exportable inside those user-defined boundaries.
- Private conversations, confidential information, and raw sensitive materials are not tradeable.
- Generalized or derived knowledge from private sources should only become tradeable when explicitly allowed by the user policy.
- `v1` should allow limited external execution only under higher trust, explicit approval, and strong policy controls.
- The default user experience should still be deliverable exchange, not execution exchange.

### Pricing, Negotiation, and Settlement

- Pricing should be mixed:
  - agents publish baseline rates and conditions,
  - real deals may negotiate from that baseline.
- Baseline pricing should use deliverable type as the anchor, with adjustments for difficulty, urgency, reputation, and SLA expectations.
- `v1` negotiation should focus on:
  - price,
  - delivery time,
  - deliverable scope and quality criteria.
- Settlement should be mixed:
  - the default is single-deliverable exchange,
  - milestone settlement appears when scope grows.
- On-chain stablecoin is the trust anchor.
- Small or recurring transactions may accrue off-chain and settle later on-chain.
- Off-chain accrual should only open after trust is established.
- Credit limits should be mixed:
  - protocol caps,
  - dynamic adjustment from history, reputation, collateral, and guarantees,
  - optional bilateral negotiation inside that envelope.

### Team Formation and Shared Artifacts

- Team formation should be mixed:
  - direct invitation is the default,
  - brief-based recruitment or bidding is available when needed.
- Role assignment should be mixed:
  - standard role templates first,
  - custom roles allowed when necessary.
- Contracting and budget control should stay centralized with the leader by default.
- Narrow downstream delegation should be possible with explicit limits.
- Shared workspace should be mixed:
  - leader-owned by default,
  - selectively shared by scope, role, and contract.
- Shared artifact flow should be mixed:
  - default experience is snapshot/package delivery,
  - live collaboration appears only under higher trust and explicit permission.

## Near-Term Sequencing

The master document should recommend the following near-term sequence:

1. finalize the top-level master document,
2. audit `External Collaboration & Economic Exchange`,
3. audit `Trust, Security & Policy`,
4. define the `P2P Knowledge Exchange` track as the first concrete product track,
5. keep the stabilization work, but move it under the master-document hierarchy,
6. follow with consolidation and UX clarity work.

## Relationship to Existing Stabilization Work

The existing stabilization plan should not be discarded. It should be reframed as:

- necessary quality work,
- subordinate to the master document,
- one execution track among several,
- guided by the product constitution rather than treated as the product strategy itself.

## Explicit Non-Goals for This Design Document

This document does not yet:

- provide a file-by-file implementation plan,
- decide final package/module boundaries,
- rewrite settings taxonomy,
- define the complete reputation formula,
- define the complete negotiation protocol,
- define the full on-chain contract model,
- or replace OpenSpec change artifacts.

Those belong in downstream execution-track documents and implementation plans.

## Success Criteria

This design is successful if the final master document lets the team answer all of the following consistently:

1. What is Lango fundamentally for?
2. Which features are central versus peripheral?
3. Which features should be stabilized, merged, deferred, or removed?
4. Which track should own each problem?
5. Why does the product start with knowledge exchange before broader execution and team coordination?
6. How do privacy, exportability, trust, and settlement relate to the product path?

