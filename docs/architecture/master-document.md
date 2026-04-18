# Lango Master Document

## Purpose

This document is the top-level product and strategy document for Lango.

It exists to:

- define what Lango fundamentally is,
- define how features are judged,
- define which execution tracks own which problems,
- keep future work aligned with one product story.

## Product Constitution

Lango is a sovereign peer-to-peer agent network, not just a local agent runtime.

### Constitutional Principles

1. External collaboration between sovereign agents is the central product identity.
2. Internal collaboration matters, but it is a reduced form of the external collaboration model.
3. External collaboration is economically native.
4. On-chain settlement is one trusted settlement mechanism, not the product itself.
5. The default external collaboration structure is leader-led.
6. The long-term product target is high autonomy, but the near-term product model is delegated autonomy.
7. User private conversations, confidential material, and sensitive internal information are never tradeable assets.
8. Tradeable knowledge and deliverables must stay inside an allowlist plus explicit exportability policy.
9. The first economic activity is knowledge exchange and result exchange, not broad remote execution.
10. Trust starts from cryptographic continuity of identity, but reputation should be centered on real collaboration and transaction history.
11. The owner carries root identity and final accountability, while individual agents accumulate separate role-specific and domain-specific reputation.

## Primary Product Path

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

## Capability Areas

Capability areas are a classification lens, not execution tracks. They describe where work belongs conceptually and can route into one or more tracks below.

### Sovereign Runtime & Operator UX

Providers, agent runtime, channels, cockpit, settings, status, session handling, logging, and core tool surfaces.

### Knowledge & Intelligence Capital

Knowledge, observational memory, embedding/RAG, graph, ontology, librarian, agent memory, skills, and context systems.

### External Collaboration & Economic Exchange

P2P identity, trust, reputation, pricing, negotiation, settlement, team formation, and shared artifact exchange.

### Trust, Security & Policy

Auth, signer/KMS, privacy boundaries, exportability policy, approvals, gatekeeper, exec safety, sandboxing, and auditability.

### Execution, Continuity & Accountability

Workflow, background tasks, cron, RunLedger, provenance, and hooks.

### Extensibility & Developer Clarity

MCP, extension packs, architecture boundaries, docs truthfulness, CLI surface clarity, and duplicate cleanup.

## Feature Audit Framework

### Allowed Audit Judgments

- `keep`
- `stabilize`
- `merge`
- `defer`
- `remove`

### Audit Rules

- A feature is valuable if it directly strengthens the core product story.
- The strongest test is whether it materially supports external collaboration, knowledge exchange, result exchange, leader-led team execution, or long-running multi-agent collaboration.
- Features with overlapping responsibility should be merge candidates.
- Features aligned with the vision but weak in reliability, defaults, docs, trust boundaries, or UX should be stabilization candidates.
- Features that fit the vision but are too early for the current product path should be deferred.
- Features that add complexity without strengthening the core story should be removal candidates.

## Track Portfolio

Tracks are execution lanes. They own concrete backlog, documents, and delivery outcomes, while capability areas stay as the stable taxonomy used to classify work.

### Stabilization Track

Truth alignment, durability, defaults, hardening, observability, and production correctness.

### Consolidation Track

Duplicate feature, config-surface, command-surface, and responsibility cleanup.

### UX Clarity Track

Settings, cockpit, CLI, docs, and defaults simplification.

### P2P Knowledge Exchange Track

The first concrete economic track for external sovereign agents.

### Leader-Led Team Execution Track

The next layer above knowledge exchange.

### Developer Clarity Track

Architecture boundary, naming, documentation, and codebase clarity improvements.

### Capability-to-Track Routing

- `Sovereign Runtime & Operator UX` routes primarily to `Stabilization Track` and `UX Clarity Track`.
- `Knowledge & Intelligence Capital` routes primarily to `Stabilization Track` and `Consolidation Track` when overlapping systems need cleanup.
- `External Collaboration & Economic Exchange` routes primarily to `P2P Knowledge Exchange Track` and, later, `Leader-Led Team Execution Track`.
- `Trust, Security & Policy` routes primarily to `Stabilization Track` when hardening or safety work is needed.
- `Execution, Continuity & Accountability` routes primarily to `Stabilization Track` and `Consolidation Track`.
- `Extensibility & Developer Clarity` is a taxonomy bucket for platform boundary work; its work routes into `Developer Clarity Track` when the goal is naming, architecture, or documentation clarity, and into `UX Clarity Track` when the goal is user-facing simplification or default cleanup.

## Near-Term Priorities

1. Audit `External Collaboration & Economic Exchange`.
2. Audit `Trust, Security & Policy`.
3. Publish the first `P2P Knowledge Exchange` track document.
4. Move stabilization work under the master-document hierarchy, meaning future stabilization docs and audits must sit under `docs/architecture/`, reference this master document as their source of truth, and declare which capability area and track they are serving.
5. Follow with consolidation and UX clarity work.
