# P2P Knowledge Exchange Track

## Document Ownership

- Primary capability area: `External Collaboration & Economic Exchange`
- Primary execution track: `P2P Knowledge Exchange Track`
- Secondary capability areas:
  - `Trust, Security & Policy`
  - `Execution, Continuity & Accountability`
- Secondary tracks:
  - `Stabilization Track`

This track charter follows `docs/architecture/master-document.md` as the top-level source of truth.

## Goal

Define the first concrete external market activity for sovereign Lango agents:

- expertise access,
- reviewable deliverables,
- bounded exportability,
- trusted settlement,
- dispute-ready receipts.

## Why This Track Comes First

Lango should reach meaningful external economic activity before it tries to support broad external execution or long-running multi-agent organizations.

Knowledge exchange is the narrowest useful slice because it:

- creates real external value,
- forces trust and exportability boundaries to become explicit,
- produces reviewable artifacts,
- creates a clean bridge toward later team execution.

## Relationship to Later Team Execution

This track is intentionally narrower than leader-led team execution.

It establishes the trust, settlement, deliverable, and exportability boundaries that later team-based collaboration depends on, without taking on full role orchestration, delegated budget control, or broader shared-workspace behavior. Leader-led team execution builds on these boundaries rather than replacing them.

The exportability policy work has already started as a first slice: source-based evaluation and operator visibility are now landed. Approval flow now has a first slice too: structured artifact release approval states and audit-backed receipts are landed. Upfront payment approval now has a first slice as well: structured decisioning, suggested payment modes, canonical transaction receipt state updates, and escrow execution input binding are landed. Dispute-ready receipts also have a lite slice now: canonical submission and transaction records, current submission pointers, and append-only event trails are in place. Direct payment execution gating is landed for the direct `prepay` path, and escrow recommendation execution is landed for the first `create + fund` path. The remaining work is deeper provenance, broader settlement progression, and dispute integration rather than starting from zero.

The first transaction-oriented runtime design slice is now documented in `docs/architecture/knowledge-exchange-runtime.md`. It ties transaction open, payment-path selection, work-start gating, submission creation, release approval, and post-approval progression into one canonical runtime story while keeping the current limits explicit.

The first settlement progression slice is now landed as well: transaction-level progression state, release-outcome mapping, review-needed handling, current-submission-gated progression writes, and the receipts-backed `apply_settlement_progression` tool are now in place. Progression updates also append to the current submission receipt event trail. `dispute-ready` remains a model-only follow-on state, and the remaining work is multi-round partial settlement, escrow lifecycle completion, and dispute engine completion.

The first direct actual settlement execution slice is now landed too: `execute_settlement` resolves canonical amount context from the transaction receipt, requires the current submission and `approved-for-settlement` state, reuses the direct payment runtime, and closes settlement progression to `settled` on success. The first direct partial settlement execution slice is now landed as well: `execute_partial_settlement` resolves canonical amount context from the transaction's `partial_settlement_hint`, requires the current submission and `approved-for-settlement` state, reuses the direct payment runtime, and closes settlement progression to `partially-settled` on success. The remaining work is escrow lifecycle completion and dispute engine completion.

The first escrow release slice is now landed too: `release_escrow_settlement` requires `escrow_execution_status = funded` plus `approved-for-settlement`, reuses the escrow runtime, and closes settlement progression to `settled` on success. The remaining work is refund, dispute-linked escrow handling, and milestone-aware release.

The first escrow refund slice is now landed too: `refund_escrow_settlement` requires `escrow_execution_status = funded` plus `review-needed`, reuses the escrow runtime, and records refund execution evidence while keeping settlement progression unchanged. The remaining work is refund terminal-state design, dispute-linked refund branching, and release-after-refund safety rules.

The first dispute hold slice is now landed too: `hold_escrow_for_dispute` requires `escrow_execution_status = funded` plus `dispute-ready`, records hold success or failure evidence, and keeps canonical escrow and settlement progression state unchanged. The remaining work is release-vs-refund adjudication, explicit held-state design, and dispute engine integration.

The first release-vs-refund adjudication slice is now landed too: `adjudicate_escrow_dispute` requires `escrow_execution_status = funded` plus `dispute-ready` and prior hold evidence, records canonical release-vs-refund branching on the transaction receipt, and leaves execution to the existing release/refund tools. The remaining work is adjudication-aware execution gating, keep-hold or re-escalation states, and broader dispute engine integration.

The first adjudication-aware release/refund execution gating slice is now landed too: adjudication success now atomically records the canonical branch and moves settlement progression, while `release_escrow_settlement` and `refund_escrow_settlement` require matching adjudication and deny on opposite-branch evidence. The remaining work is automatic post-adjudication execution, keep-hold or re-escalation states, and broader dispute engine integration.

The first automatic post-adjudication execution slice is now landed too: `adjudicate_escrow_dispute` accepts optional `auto_execute=true`, keeps adjudication as the canonical write layer, and may inline the matching release or refund executor while still reusing the same executor gates. The remaining work is background execution, retry orchestration, automatic execution as policy default, and broader dispute engine integration.

The first background post-adjudication execution slice is now landed too: `adjudicate_escrow_dispute` accepts optional `background_execute=true`, enqueues the canonical release or refund follow-up onto the existing background task substrate, and returns a dispatch receipt while leaving actual execution asynchronous. The remaining work is retry orchestration, dead-letter handling, dedicated status observation, and policy-driven defaults.

The first retry / dead-letter slice is now landed too: background post-adjudication execution now retries up to `3` times with exponential backoff, tracks retry metadata on background tasks, and appends retry scheduled / dead-lettered evidence without changing canonical adjudication. The remaining work is operator replay, generic async retry policy, dead-letter browsing, and policy-driven backoff tuning.

The first operator replay / manual retry slice is now landed too: `retry_post_adjudication_execution` requires dead-letter evidence plus canonical adjudication, appends `manual-retry-requested` evidence, and creates a fresh background post-adjudication dispatch without clearing prior dead-letter evidence. The remaining work is dead-letter browsing UI, policy-driven replay controls, generic replay substrate design, and broader dispute engine integration.

The first dead-letter browsing / status observation slice is now landed too: operators can list the current dead-lettered post-adjudication backlog and inspect the canonical snapshot plus latest retry/dead-letter summary for a given transaction. The remaining work is richer filtering, raw background-task bridges, and higher-level cockpit or CLI surfaces.

## In Scope

- pseudonymous but cryptographically continuous identities,
- owner-root trust plus agent-specific reputation,
- allowlist plus explicit exportability policy,
- structured artifact release approval and audit-backed release receipts,
- structured upfront payment approval decisioning and receipt updates,
- receipt-backed direct payment execution gating for the direct `prepay` path,
- receipt-backed escrow recommendation execution for the first `create + fund` path,
- partial settlement execution,
- expertise access and reviewable deliverables,
- small upfront payment plus approval-based final settlement,
- on-chain stablecoin as the trust anchor,
- off-chain accrual only after trust is established,
- dispute handling grounded in signed logs, provenance, acceptance criteria, escrow state, and immutable receipts.

## Out of Scope

- unrestricted remote execution,
- full team-based role orchestration,
- live shared workspaces by default,
- complete reputation formula design,
- final smart-contract design,
- human approval UI,
- dispute orchestration,
- escrow lifecycle completion.

## Default Transaction Shape

1. A leader agent discovers or selects an external counterparty.
2. The leader agent confirms the target artifact is tradeable under exportability policy.
3. The parties agree on price, delivery window, and deliverable scope.
4. A small upfront payment or an escrow can now be created for the first landed payment paths. The escrow path currently covers `create + fund` only.
5. The external agent delivers a reviewable artifact.
6. The leader agent approves, rejects, requests revision, escalates, or disputes the artifact.
7. Final settlement is released on approval, partially settled when the canonical partial-settlement hint applies, deferred for revision or escalation, or handled through dispute rules.

## Required Follow-On Plans

1. `identity / trust / reputation` detailed audit is now landed; the follow-on work is `reputation v2`, stronger trust-entry contracts, and runtime integration
2. `pricing / negotiation / settlement` detailed audit is now landed; the follow-on work is runtime integration, refund terminal-state design, and broader dispute completion
3. exportability policy follow-on work (the first source-primary slice has landed; the remaining gaps are richer rules, override/dispute handling, and receipt unification)
4. `settlement progression` first slice is now landed; the follow-on work is partial settlement rules, dispute engine completion, and deeper disagreement handling
5. `actual settlement execution` first slice is now landed; `partial settlement execution` first slice is now landed too; the follow-on work is dispute-linked escrow handling and deeper settlement orchestration
6. `escrow release` first slice is now landed; the follow-on work is refund, dispute-linked escrow handling, and milestone-aware release
7. `escrow refund` first slice is now landed; the follow-on work is refund terminal-state design, dispute-linked refund branching, and release-after-refund safety rules
8. `dispute hold` first slice is now landed; the follow-on work is release-vs-refund adjudication, explicit held-state design, and dispute engine integration
9. `release vs refund adjudication` first slice is now landed; the follow-on work is adjudication-aware release/refund execution, keep-hold or re-escalation states, and broader dispute engine integration
10. `adjudication-aware release/refund execution gating` first slice is now landed; the follow-on work is automatic post-adjudication execution, keep-hold or re-escalation states, and broader dispute engine integration
11. `automatic post-adjudication execution` first slice is now landed; the follow-on work is background execution, retry orchestration, automatic execution as policy default, and broader dispute engine integration
12. `background post-adjudication execution` first slice is now landed; the follow-on work is retry orchestration, dead-letter handling, dedicated status observation, and policy-driven defaults
13. `retry / dead-letter handling` first slice is now landed; the follow-on work is operator replay, generic async retry policy, dead-letter browsing, and policy-driven backoff tuning
14. `operator replay / manual retry` first slice is now landed; the follow-on work is dead-letter browsing UI, policy-driven replay controls, generic replay substrate design, and broader dispute engine integration
15. `dead-letter browsing / status observation` first slice is now landed; the follow-on work is richer filtering, raw background-task bridges, and higher-level cockpit or CLI surfaces
16. the first transaction-oriented runtime design slice, now documented in `docs/architecture/knowledge-exchange-runtime.md`; follow-on work is runtime implementation and broader progression handling
