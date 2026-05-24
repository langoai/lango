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

The current settlement progression slice is now landed as well: transaction-level progression state, release-outcome mapping, current-submission-gated progression writes, and the receipts-backed `apply_settlement_progression` tool are now in place. `escalate` now re-enters canonical `dispute-ready` when disagreement returns from `review-needed`, `approved-for-settlement`, or `partially-settled`, preserved `partial_settlement_hint` survives that re-entry, and tool receipts now expose `dispute_lifecycle_status` alongside the progression fields. The remaining work is repeated partial execution, broader multi-round settlement orchestration, and operator/policy surfaces rather than inventing a second canonical progression model.

The first direct actual settlement execution slice is now landed too: `execute_settlement` resolves canonical amount context from the transaction receipt, requires the current submission and `approved-for-settlement` state, reuses the direct payment runtime, and closes settlement progression to `settled` on success. The first direct partial settlement execution slice is now landed as well: `execute_partial_settlement` resolves canonical amount context from the transaction's `partial_settlement_hint`, requires the current submission and `approved-for-settlement` state, reuses the direct payment runtime, and closes settlement progression to `partially-settled` on success. The remaining work is escrow lifecycle completion and dispute engine completion.

The first escrow release slice is now landed too: `release_escrow_settlement` requires `escrow_execution_status = funded`, `approved-for-settlement`, and matching `escrow_adjudication = release`, reuses the escrow runtime, blocks opposite-branch refund evidence, and clears the active dispute lifecycle marker when release closes progression to `settled`. The remaining work is milestone-aware release, broader execution policy defaults, and richer operator policy surfaces.

The first escrow refund slice is now landed too: `refund_escrow_settlement` requires `escrow_execution_status = funded`, `review-needed`, and matching `escrow_adjudication = refund`, reuses the escrow runtime, blocks opposite-branch release evidence, records refund execution evidence, clears the active dispute lifecycle marker while keeping settlement progression unchanged, and now serializes concurrent refund attempts per transaction inside the service boundary. The remaining work is refund terminal-state design, release-after-refund safety rules, and richer operator policy surfaces.

The first `reputation v2 + runtime integration` slice is now landed too: `internal/p2p/reputation` now separates composite compatibility score, earned trust score, durable negative units, and temporary safety signals; canonical trust entry now resolves `bootstrap`, `established`, `review`, and `temporarily_unsafe`; firewall admission, handshake auto-approval, economy pricing/risk wiring, post-pay routing, and the team reputation bridge now consume that contract instead of inventing local trust meanings. The remaining work is owner-root-aware policy adoption, broader dispute-to-reputation feeds, and richer operator-facing trust surfaces rather than another reputation redesign.

The first dispute hold slice is now landed too: `hold_escrow_for_dispute` requires `escrow_execution_status = funded` plus `dispute-ready`, records hold success or failure evidence, keeps canonical settlement progression at `dispute-ready`, now sets `dispute_lifecycle_status = hold-active` on success, and serializes concurrent hold attempts per transaction inside the service boundary. Tool receipts also surface that lifecycle state directly. The remaining work is richer arbitration policy, a separate held-escrow lifecycle only if it proves necessary, and operator UI.

The first release-vs-refund adjudication slice now also serializes concurrent adjudication attempts per transaction inside the service boundary so the canonical branch update is not applied in parallel.

The first release-vs-refund adjudication slice is now landed too: `adjudicate_escrow_dispute` requires `escrow_execution_status = funded` plus `dispute-ready` and prior hold evidence, records canonical release-vs-refund branching on the transaction receipt, moves settlement progression atomically onto the matching branch, preserves the active dispute lifecycle marker, and by default stops at canonical adjudication for manual recovery unless an execution flag is set. Tool receipts now surface `dispute_lifecycle_status` directly. The remaining work is config-backed non-manual defaults, richer arbitration policy, and operator UI.

The first adjudication-aware release/refund execution gating slice is now landed too: adjudication success now atomically records the canonical branch and moves settlement progression, while `release_escrow_settlement` and `refund_escrow_settlement` require matching adjudication, deny on opposite-branch evidence, and clear dispute lifecycle state only when terminal execution succeeds. The remaining work is milestone-aware branch execution, broader dispute automation, and operator/policy surfaces.

The first automatic post-adjudication execution slice is now landed too: `adjudicate_escrow_dispute` accepts optional `auto_execute=true`, keeps adjudication as the canonical write layer, and may inline the matching release or refund executor while still reusing the same executor gates. Post-adjudication follow-up now resolves through one execution-mode policy: `auto_execute=true` selects inline execution, `background_execute=true` selects background execution, and omitted execution flags default to manual recovery. The remaining work is config-backed non-manual defaults, policy editing for execution-mode selection, and broader dispute engine integration.

The first background post-adjudication execution slice is now landed too: `adjudicate_escrow_dispute` accepts optional `background_execute=true`, enqueues the canonical release or refund follow-up onto the existing background task substrate, and returns a dispatch receipt while leaving actual execution asynchronous. This background mode is now one branch of the shared post-adjudication execution policy rather than a standalone surface-specific path. The remaining work is config-backed non-manual defaults, operator-editable execution-mode policy, broader background-task adoption outside post-adjudication follow-up, and broader dispute engine integration.

The first retry / dead-letter slice is now landed too: background post-adjudication execution now retries up to `3` times with exponential backoff from a normalized runtime retry policy shape, tracks retry metadata on background tasks, appends `post_adjudication_retry` evidence with `retry-scheduled` and `dead-lettered` subtypes, preserves canonical adjudication, and now re-escalates settlement progression back to `dispute-ready` with `dispute_lifecycle_status = re-escalated` when retries exhaust. The remaining work is operator-editable retry tuning, wider non-post-adjudication adoption of the retry policy shape, and a more generic recovery substrate for arbitrary background task families.

The first operator replay / manual retry slice is now landed too: `retry_post_adjudication_execution` requires dead-letter evidence plus canonical adjudication, appends `manual-retry-requested` evidence, and creates a fresh background post-adjudication dispatch without clearing prior dead-letter evidence. Replay now uses the same `post_adjudication_retry` evidence family as automatic retry and dead-letter handling, so manual recovery sits on the same recovery substrate as background recovery. The remaining work is inline replay, arbitrary background-task replay, per-transaction recovery snapshots, and broader dispute engine integration.

The first policy-driven replay controls slice is now landed too: replay now resolves the current actor from runtime context, applies config-backed allowlists for overall replay plus outcome-specific replay, and fails closed when actor resolution or authorization fails. These allowlists now sit on top of the shared recovery evidence gate rather than a replay-only side path. The remaining work is richer policy classes, policy editing surfaces, per-transaction snapshots, and amount-tier replay controls.

The dead-letter browsing / status observation slice now includes transaction-global dominant family, compact per-submission breakdown, a thin raw background-task bridge on the detail view, a first cockpit read surface, a first cockpit filter bar, subtype filtering, latest-family filtering, any-match-family filtering, actor/time filtering, reason/dispatch filtering, a first cockpit recovery action with confirm/refresh UX, richer retry loading/failure feedback, reset/clear shortcuts, selection preservation, and a dedicated dead-letter CLI surface too. Operators can filter the current dead-lettered post-adjudication backlog by adjudication outcome, receipt-ID query, latest manual replay actor, latest dead-letter time window, dead-letter reason substring, latest dispatch reference, latest status subtype, latest subtype family, and any-match family from the dedicated CLI surface, while the broader list tool still exposes the deeper retry-count and transaction-global family filters. Each row now also includes a compact `submission_breakdown` over all submissions in the transaction, and the detail view exposes an optional `latest_background_task` with the latest matching task ID, status, attempt count, and next retry time. Cockpit offers a read-only master-detail dead-letter surface with `query`, `adjudication`, `latest_status_subtype`, `latest_status_subtype_family`, `any_match_family`, `manual_replay_actor`, `dead_lettered_after/before`, `dead_letter_reason_query`, and `latest_dispatch_reference` filtering, and all of those filter fields are now forwarded through the cockpit bridge into the dead-letter list tool instead of being truncated at the shell adapter. A `Ctrl+R` full filter reset clears draft/applied state and retry confirm state, unified selection preservation spans apply, reset, and retry-success refresh, and retry messaging moves through confirm, running, request-accepted follow-up, and failure states while distinguishing retry-request acceptance from invocation failure. Both CLI and cockpit retry paths now inject a local default operator principal when the runtime context would otherwise be empty, so replay policy evaluates a concrete actor instead of failing immediately on missing principal. The page-top compact summary strip is driven from the current backlog rows with total/retryable/adjudication/latest-family overview plus raw top latest dead-letter reasons on a compact `reasons:` line, grouped latest reason-family buckets on a compact `reason families:` line, top latest manual replay actors on an `actors:` line, grouped latest manual replay actor families on an `actor families:` line, grouped latest dispatch-reference families on a `dispatch families:` line, raw top latest dispatch references on a `dispatch:` line, and a compact trend line over recent backlog timestamps. The initial reason-family taxonomy is `retry-exhausted`, `policy-blocked`, `receipt-invalid`, `background-failed`, and `unknown`, using case-insensitive heuristics over each current `latest_dead_letter_reason` with `unknown` fallback; raw top latest dead-letter reasons remain available alongside the grouped view. The initial actor-family taxonomy is `operator`, `system`, `service`, and `unknown`, using case-insensitive heuristics over each current `latest_manual_replay_actor` with `unknown` fallback; raw top latest manual replay actors remain available alongside the grouped view. Dispatch-family grouping now uses the same shared classifier in both CLI and cockpit: a compact prefix classifier over each current `latest_dispatch_reference` that recognizes common families such as `dispatch`, `queue`, `worker`, `bridge`, `webhook`, and `unknown`, normalizes `job` / `runner` / `task` to `worker`, and otherwise preserves the first normalized token. The CLI now also exposes `lango status dead-letter-summary` for the backlog overview surface with total dead letters, retryable count, adjudication buckets, latest-family buckets, grouped `by_reason_family`, `by_actor_family`, and `by_dispatch_family` buckets, raw configurable top-N latest reasons, actors, and dispatch references, and a recent trend window controlled by `--top`, `--trend-window`, and `--trend-bucket`. `lango status dead-letters` now includes any-match-family parity with the cockpit surface, while `lango status dead-letter retry <transaction-receipt-id>` can emit an immediate structured follow-up snapshot and optionally poll for follow-up change with `--wait`, `--wait-interval`, and `--wait-timeout`. The remaining work is configurable taxonomy redesign, broader dead-letter history and generic background-task browsing, wider non-post-adjudication adoption of the retry/recovery substrate, and operator-editable execution/recovery policy surfaces.

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

1. `identity / trust / reputation` detailed audit is now landed, and the first `reputation v2 + runtime integration` slice is now landed too; the follow-on work is owner-root-aware policy adoption, broader dispute-to-reputation feeds, and richer operator-facing trust/review surfaces
2. `pricing / negotiation / settlement` detailed audit is now landed; the follow-on work is runtime integration, repeated partial execution, and broader dispute/recovery policy completion
3. exportability policy follow-on work (the first source-primary slice has landed; the remaining gaps are richer rules, override/dispute handling, and receipt unification)
4. `settlement progression` current slice is now landed with explicit dispute-ready re-entry and partial-hint preservation; the follow-on work is repeated partial execution, broader multi-round settlement orchestration, and operator/policy surfaces
5. `actual settlement execution` first slice is now landed; `partial settlement execution` first slice is now landed too; the follow-on work is repeated partial execution, milestone-aware escrow branches, and broader recovery/policy orchestration
6. `escrow release` first slice is now landed with adjudication gating; the follow-on work is milestone-aware release, broader execution policy defaults, and richer operator policy surfaces
7. `escrow refund` first slice is now landed with adjudication gating; the follow-on work is refund terminal-state design, release-after-refund safety rules, and richer operator policy surfaces
8. `dispute hold` first slice is now landed with canonical `hold-active` lifecycle state; the follow-on work is richer arbitration policy, a separate held-escrow lifecycle only if needed, and operator UI
9. `release vs refund adjudication` first slice is now landed with manual-recovery-by-default canonical branching and lifecycle-state tool receipts; the follow-on work is config-backed non-manual defaults, richer arbitration policy, and operator UI
10. `adjudication-aware release/refund execution gating` first slice is now landed; the follow-on work is milestone-aware branch execution, broader dispute automation, and operator/policy surfaces
11. `automatic post-adjudication execution` first slice is now landed; the follow-on work is config-backed non-manual defaults, policy editing for execution-mode selection, and broader dispute engine integration
12. `background post-adjudication execution` first slice is now landed; the follow-on work is config-backed non-manual defaults, operator-editable execution-mode policy, broader background-task adoption outside post-adjudication follow-up, and broader dispute engine integration
13. `retry / dead-letter handling` first slice is now landed with canonical re-escalation on exhausted retries, canonical retry-key dedup across pending/running/scheduled tasks, and explicit panic-to-failed-task handling in the background manager; the follow-on work is operator-editable retry tuning, wider non-post-adjudication adoption of the retry policy shape, and a more generic recovery substrate for arbitrary background task families
14. `operator replay / manual retry` first slice is now landed; the follow-on work is inline replay, arbitrary background-task replay, per-transaction recovery snapshots, and broader dispute engine integration
15. `policy-driven replay controls` first slice is now landed; the follow-on work is richer policy classes, policy editing surfaces, per-transaction snapshots, and amount-tier replay controls
16. `dead-letter browsing / status observation` slice now includes local and transaction-global family grouping, compact per-submission breakdown, a thin raw background-task bridge on the detail view, a cockpit read surface, a thin cockpit filter bar with subtype, latest-family, any-match-family, actor/time, and reason/dispatch filtering, a full `Ctrl+R` reset shortcut, unified selection preservation, a replay/write control with confirm/refresh plus refined running/failure/success feedback UX, a richer dead-letter CLI summary surface with grouped `by_reason_family`, `by_actor_family`, and `by_dispatch_family` buckets while preserving raw configurable top-N latest reasons, actors, and dispatch references plus a recent trend window, and a dead-letter CLI surface with latest-subtype / latest-family / any-match-family plus actor/time and reason/dispatch list filtering, `offset/limit` pagination, explicit `--actor` retry override, retry follow-up polling semantics, and structured JSON error payloads for machine-mode failures; the follow-on work is configurable taxonomy redesign, broader dead-letter history and generic background-task browsing, wider non-post-adjudication adoption of the retry/recovery substrate, and operator-editable execution/recovery policy surfaces
17. the first transaction-oriented runtime design slice, now documented in `docs/architecture/knowledge-exchange-runtime.md`; follow-on work is runtime implementation and broader progression handling
