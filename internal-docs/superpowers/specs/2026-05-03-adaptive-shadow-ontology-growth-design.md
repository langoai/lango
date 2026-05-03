# Adaptive Shadow Ontology Growth Design

## Context

Lango's knowledge, learning, ontology, graph, and FTS5 systems are meant to support an agent that can collect, classify, organize, retrieve, and evolve knowledge over time. The runtime already contains most of the ontology lifecycle primitives needed for that goal, including predicate lifecycle statuses, graph buffering, truth maintenance, entity alias resolution, and governance configuration.

The current failure mode is not the absence of ontology infrastructure. The problem is that unknown predicates discovered by LLM-driven producers reach existing validation gates before they are normalized, proposed, shadowed, or quarantined. As a result, semantically useful discoveries such as `includes` or `defines` are treated as graph write errors instead of schema growth signals.

## Problem Statement

Unknown ontology terms are currently rejected too late and at multiple boundaries:

- `OntologyService.ValidateTriple` rejects unknown or deprecated predicates.
- `graph.BoltStore.putTriple` rejects unknown predicates when a validator is injected.
- Several producers and tool paths write directly to `GraphBuffer`, `graph.Store.AddTriples`, or `AssertFact` without a single admission path.

This creates four runtime failures:

- Background knowledge growth can fail noisily even when the user-facing response succeeds.
- LLM-discovered schema extensions are treated as hard errors instead of governance candidates.
- One invalid triple can roll back an entire BoltDB batch transaction.
- The current behavior depends on which producer path emitted the triple.

## Goals

- Reuse and extend existing ontology and graph assets instead of duplicating them.
- Introduce a single triple admission boundary for all dynamic or untrusted triple producers before unknown-predicate validation decisions.
- Preserve user-response continuity even when ontology growth or graph storage fails.
- Convert unknown predicates and types into mapped, proposed, shadow, quarantined, or dead-lettered outcomes.
- Keep graph predicate validation as a final integrity guard, but feed it from one authoritative predicate state.
- Define the first-wave retrieval, FTS5, governance, and replay semantics for shadow predicates.

## Non-Goals

- Do not replace the existing ontology lifecycle FSM.
- Do not remove `ValidateTriple` or graph store predicate validation.
- Do not invent a second independent dead-letter subsystem unrelated to existing replay/status patterns.
- Do not assume automatic active promotion is already implemented everywhere because config and docs mention it.

## Existing Assets And Change Scope

| Asset | Exists Today | Role In This Design | Change Scope |
|---|---|---|---|
| `PredicateDefinition.Status` with `proposed/quarantined/shadow/active/deprecated` | Yes | Reused as the only schema lifecycle model | No new status model; admission feeds existing statuses |
| `OntologyService.RegisterPredicate`, `PromotePredicate`, `PredicateValidator` | Yes | Reused as authoritative schema and validation API | Extend usage, not replace API |
| `CandidateTriple` in truth maintenance | Yes | Stays as conflict snapshot model | No reuse for ingestion; new admission candidate type is separate and narrower in purpose |
| `AliasStore`, `DeclareSameAs`, `Merge`, `Split` | Yes | Reused for entity canonicalization patterns | Predicate aliasing may extend the pattern, but does not replace entity alias storage |
| `GraphBuffer` | Yes | Remains the batch write stage after admission | Inputs change; buffer remains the batcher |
| `graph.BoltStore` validator injection | Yes | Remains final store-level guard | No independent predicate state inside store; it consumes the ontology validator closure |
| Existing dead-letter UI and replay surfaces | Yes | Reused as architectural pattern for inspect/replay UX | New admission dead letters may reuse storage/replay conventions, not existing post-adjudication event records |
| `GovernancePolicy` fields such as `MaxNewPerDay`, `SchemaExplosionBudget`, `MinUsageForPromotion` | Yes | Reused as policy knobs | Runtime enforcement must be clarified and extended where not yet implemented |

## Recommended Policy

Use an `adaptive-shadow` policy, but ground it in existing ontology governance rather than a new parallel schema system.

Unknown ontology terms move through one deterministic admission sequence:

1. Resolve whether the predicate or type is already `active` or `shadow`.
2. Attempt canonical mapping to an existing predicate.
3. Evaluate confidence and source evidence.
4. Check governance and schema growth budgets.
5. Decide one of `known`, `mapped`, `shadow`, `proposed`, `quarantined`, or `dead-lettered`.

Decision meanings:

- `known`: existing `active` or `shadow` schema entry; admit unchanged.
- `mapped`: rewrite to a canonical predicate and admit the rewritten triple.
- `shadow`: move a new predicate/type into existing `shadow` lifecycle state, then admit.
- `proposed`: persist the candidate and evidence, but do not admit the triple to the graph.
- `quarantined`: persist the candidate and rejection reason because it is too generic, conflicting, or explosive.
- `dead-lettered`: admission already decided the triple should be admitted, but a later storage or consistency step failed.

`shadow` continues to mean "usable but experimental." The design does not create a second shadow concept.

Decision table for non-admitted unknown terms in Wave 1:

| Condition | Outcome |
|---|---|
| low confidence, first-seen, or insufficient repeated evidence | `proposed` |
| budget exhausted but candidate is otherwise plausible | `proposed` |
| generic predicate heuristic hit | `quarantined` |
| conflicting semantic mapping or suspicious source/evidence | `quarantined` |
| storage or consistency failure after admission decision | `dead-lettered` |

## Authoritative Validation Model

There must be one source of truth for predicate validity.

- The authoritative predicate state is the ontology service predicate cache behind `PredicateValidator()`.
- `ValidateTriple` and the graph store validator must both consume that same predicate truth.
- The graph store must not maintain a separate predicate set.
- "Refresh predicate validator immediately" means refreshing the ontology service cache and continuing to use the same closure already injected into `BoltStore`.
- In the current runtime, `RegisterPredicate` already performs a synchronous `refreshPredicateCache()` before returning; first-wave admission relies on that synchronous invalidation model rather than a separate publish/subscribe refresh path.

This matters because the current runtime rejects unknown predicates in two places. The design is valid only if both gates read from the same refreshed ontology state.

Important Wave 1 constraint:

- With governance enabled, `RegisterPredicate` forces new predicates to `proposed` status before persistence.
- Therefore the shadow admission path cannot rely on a single `RegisterPredicate(Status=shadow)` call.
- Wave 1 shadow admission must use one of two explicit strategies:
  - `RegisterPredicate(proposed)` followed by `PromotePredicate(proposed -> shadow)` in the same admission workflow.
  - a future internal-only API that preserves shadow registration semantics without exposing a second public lifecycle surface.
- The default design assumption for Wave 1 is the existing-API two-step path: register as `proposed`, then promote to `shadow`, then admit only after the shared validator closure recognizes the refreshed shadow predicate.

## Producer Call-Site Inventory

The admission boundary is only meaningful if it covers the real producer set.

| Producer Path | Current Entry | Current Validation Path | Confidence Available Today | First-Wave Action |
|---|---|---|---|---|
| `internal/learning/parse.go` -> `TriplesExtractedEvent` -> `internal/app/wiring_graph.go` | Event bus to `GraphBuffer` | Graph store validator only | No explicit numeric confidence on event triple | Route through admission in event subscriber; add confidence/source metadata or default producer policy |
| `internal/librarian/proactive_buffer.go` -> `TriplesExtractedEvent` | Event bus to `GraphBuffer` | Graph store validator only | No explicit numeric confidence on event triple | Same as learning event path |
| `internal/learning/graph_engine.go` direct graph writes | `graph.Store.AddTriples` | Graph store validator only | Producer-local knowledge of source exists; no shared admission shape | Reroute through admission batch API before direct store writes |
| `internal/ontology/truth.go` via `AssertFact` | `ValidateTriple` before truth maintenance | Ontology validator first | Yes, `AssertionInput.Confidence` exists today | Reuse confidence; route schema discovery before `ValidateTriple` hard rejection |
| `internal/app/wiring_graph.go` content saved containment triples | `GraphBuffer` | Known seeded predicates | Deterministic | Wave 1 keeps a fast path and bypasses dynamic admission because this producer only emits seeded predicates; existing validators still apply unchanged |
| `internal/memory/graph_hooks.go` | `GraphBuffer` | Known seeded predicates | Deterministic | Wave 1 keeps a fast path and bypasses dynamic admission because this producer only emits seeded predicates; existing validators still apply unchanged |
| `internal/cli/graph/import_cmd.go` | Direct `AddTriples` | Graph store validator only | Import payload dependent | Route import through admission batch API |
| Ontology tools and actions using `AssertFact` | `OntologyService.AssertFact` | `ValidateTriple` first | Tool input controlled | Admission must sit before unknown-predicate rejection for growth-enabled flows |
| Type-bearing dynamic producers | Same as the corresponding dynamic triple path | Same as predicate path | Depends on producer | Unknown type discovery follows the same admission path whenever the producer emits non-empty `SubjectType` or `ObjectType` values |

The key boundary decision is this: Wave 1 keeps a fast path for deterministic seeded-predicate producers and requires the admission API for every path that can emit LLM- or user-supplied arbitrary predicates. This does not weaken the validation model because both fast-path and admission-path triples still terminate at the same existing ontology and graph validators.

`internal/app/wiring_graph.go` contains both sides of this split today: it has dynamic subscriber paths that ingest `TriplesExtractedEvent` values, and it also emits deterministic containment triples locally. Wave 1 applies admission only to the dynamic subscriber side in that file.

## Admission Architecture

This is an extend-not-duplicate design.

```text
raw producer
  -> TripleAdmissionPolicy
  -> ontology-backed decision
  -> admitted triples only
  -> existing GraphBuffer
  -> existing graph.Store
```

### Components

`AdmissionCandidate`

New narrow input model for graph admission. This is not a replacement for truth-maintenance `CandidateTriple`.

Field split:

- `AdmissionCandidate` is an ingestion model and carries producer, session key, turn ID, optional types, source, confidence, and evidence metadata.
- truth-maintenance `CandidateTriple` is a conflict snapshot and carries only subject, object, source, confidence string, and recorded timestamp for contradiction reporting.

`TripleAdmissionPolicy`

New single entry point for graph-bound triples. It owns the top-level admission decision, batch pre-filtering, canonical mapping, and routing to proposal/quarantine/dead-letter outcomes.

`OntologyGrowthEngine`

New helper used only by `TripleAdmissionPolicy` on the unknown-predicate or unknown-type branch. It does not act as a second entrypoint and does not own lifecycle state. Its role is to execute the unknown-branch growth workflow by calling existing ontology service operations or candidate persistence.

`SchemaCandidateStore`

New or extended persistence for schema candidates and evidence. Wave 1 keeps this deliberately small: one candidate row per normalized unknown term, recent `K=20` sample admissions, per-source counters, subject/object type frequency summaries, last rejection reason, and last observed schema version.

`GraphAdmissionDeadLetterStore`

New admission-domain dead-letter persistence. This is not the existing post-adjudication event store. It should reuse the existing dead-letter UI and replay pattern shape where useful, but it stores graph admission failures, not post-adjudication execution failures.

`SchemaBudgetLedger`

Required extension for budget enforcement if automatic shadow admission is enabled. Existing `GovernancePolicy` fields already define daily and monthly budgets, but current governance runtime only has an in-memory daily counter. Persisted budget accounting is needed before production auto-shadow mode. Responsibility remains one-way: `TripleAdmissionPolicy` invokes `OntologyGrowthEngine`, and `OntologyGrowthEngine` consumes and updates the budget ledger when evaluating growth decisions.

## Batch And Transaction Model

First wave chooses one batch model explicitly:

- Admission pre-filters and rewrites candidates before they reach `GraphBuffer`.
- `GraphBuffer` continues to submit one atomic BoltDB transaction per admitted batch.
- Invalid or deferred candidates are removed before batch write.
- The system does not split normal batches into per-triple transactions in first wave.

This means the primary fix for batch rollback is not retry fan-out. The fix is that `GraphBuffer` no longer receives raw unknown predicates from dynamic producers.

Dead letters therefore narrow to two cases:

- admission accepted a triple but later storage failed;
- ontology state changed or was inconsistent between admission and storage.

## Data Flow

Canonical mapping path:

```text
raw triple: project includes module
  -> admission sees unknown predicate
  -> predicate alias policy maps includes -> contains
  -> admitted triple uses contains
  -> existing GraphBuffer writes batch atomically
```

Shadow growth path:

```text
raw triple: term defines concept
  -> admission sees unknown predicate
  -> confidence and evidence clear governance checks
  -> existing ontology service registers defines as proposed
  -> existing ontology service promotes defines to shadow
  -> ontology predicate cache refreshes on both lifecycle writes
  -> same validator closure now recognizes defines
  -> admitted triple reaches GraphBuffer
```

Proposed-only path:

```text
raw triple: thing uses stuff
  -> admission sees unknown generic predicate
  -> candidate evidence stored
  -> predicate remains proposed or quarantined
  -> triple never enters GraphBuffer
  -> user response path continues
```

## Retrieval And FTS5 Semantics

The current FTS5 indexes are attached to knowledge and learning stores, not to graph triples directly. That means shadow predicate behavior must be defined in graph retrieval terms, not by pretending predicates create direct FTS5 rows.

First-wave semantics:

- Shadow predicates are usable for graph storage and graph traversal because they are valid ontology predicates in `shadow` state.
- GraphRAG may traverse shadow predicates, but retrieval results expanded through shadow predicates must be marked experimental in provenance metadata and should apply a score penalty relative to `active` predicate edges.
- Knowledge-store FTS5 and learning-store FTS5 behavior does not change solely because a predicate is shadow. Existing FTS5 sync stays tied to knowledge and learning writes.

This preserves the meaning of "usable but experimental" without falsely implying that shadow predicates are identical to active predicates in retrieval.

## Governance And Promotion Boundaries

The design must distinguish between existing policy surface and verified runtime behavior.

Existing surface already includes:

- lifecycle statuses;
- governance config for daily and monthly budgets;
- `ontology_promote_predicate`;
- documented `minUsageForPromotion` and shadow duration fields.

First-wave design assumptions:

- manual promotion through existing governance surfaces is the reliable promotion path;
- automatic active promotion should not be a prerequisite for this admission change;
- if auto-promotion runtime is incomplete or unverified, the change must state that shadow-to-active automation is a follow-up wave.

Therefore:

- Wave 1 uses manual promotion plus status/usage observability.
- Wave 2 may implement or harden automatic shadow-to-active promotion using the already defined `MinUsageForPromotion` policy and shadow usage counters.

## Observe-Only Phase

Change A Phase A1 is strictly non-enforcing.

| Behavior | Phase A1 |
|---|---|
| compute admission decision | yes |
| emit metrics and structured logs | yes |
| rewrite predicates via canonical mapping | no |
| register or promote schema entries | no |
| write admission dead letters | no |
| change existing graph write routing | no |
| reduce current unknown-predicate `ERROR` logs | no |

This means the operational acceptance criteria about error-rate reduction apply after Phase A2 enforcement, not after Phase A1 observe-only instrumentation.

## Budget Enforcement

Existing `GovernancePolicy` already defines `MaxNewPerDay` and `SchemaExplosionBudget`, but the current governance engine only persists an in-memory daily count per process restart boundary.

The design therefore requires:

- UTC-day keyed daily accounting for proposal intake;
- UTC-month keyed accounting for schema explosion budget;
- a persisted ledger if auto-shadow is enabled beyond observe-only mode;
- explicit behavior for multi-process or restarted runtimes.

Boundary rule:

- observe-only and proposed-only modes may temporarily reuse process-local counters;
- auto-shadow admission must not rely only on process-local counters.

## Replay And Idempotency

Admission dead-letter replay must remain auditable when schema changes between failure and replay.

Each dead-letter record must store:

- the original admission candidate;
- the schema version observed during admission;
- the decision snapshot, if any;
- the failure stage and error.

Replay rules:

- the current `SchemaVersion()` value is an in-memory atomic counter and is not durable across process restarts;
- therefore Wave 1 and Wave 2 may only use schema version as a best-effort in-process signal, not as durable replay truth;
- if schema version is unchanged within the same process lifetime, replay should attempt the same admission decision first;
- if schema version advanced within the same process lifetime, replay reruns admission against the latest schema but preserves the original snapshot for audit;
- durable replay semantics across restart require a persisted schema version source and belong to Change C prerequisite work;
- replay deduplication uses candidate hash plus the best available schema-version snapshot to avoid repeated duplicate admissions.

## Generic Predicate Policy

The generic predicate denylist must be operational, not an undocumented hardcoded bag of strings.

First-wave rule:

- generic-predicate heuristics live in one policy module with config-backed defaults;
- the Wave 1 config lives under the `ontology.governance` namespace as admission-policy fields, not as ontology rows;
- the initial defaults may include values such as `has`, `is`, `uses`, `related`, `thing`, and `misc`;
- additions and removals must be configuration-driven rather than scattered hardcoded checks;
- the policy must not assume English-only future inputs, even if the first default list is English.

## Out-Of-Scope Follow-Ups

- shadow-aware lexical search over graph predicates, because current FTS5 indexes do not index graph edges as first-class rows;
- full automatic shadow-to-active promotion if current runtime hardening is incomplete.

## Error Handling

Unknown predicate discovery is not an application error by default.

- unknown predicates handled by admission should produce structured schema-growth events, not graph write `ERROR`s;
- validator/cache mismatch after successful shadow registration is an `ERROR`;
- graph write failure after successful admission is a dead-letter event;
- user response generation is independent from admission and graph write outcomes.

## Observability

The system should expose:

- unknown predicate and type discovery counts;
- canonical mapping counts;
- shadow creation counts;
- proposed/quarantined/dead-letter counts;
- top schema candidates by frequency;
- shadow edge retrieval counts;
- graph admission rejection reasons;
- budget ledger consumption;
- validator/cache mismatch count.

These should land in internal status surfaces before any public doc claims are expanded.

## Testing Strategy

Unit tests:

- `includes` maps to `contains`;
- `defines` can complete the governance-aware `proposed -> shadow` path and then pass the same validator closure;
- low-confidence unknown predicates become proposed-only;
- generic predicates, suspicious sources, and conflicting mappings route to the expected proposed/quarantined outcomes;
- generic predicates are quarantined by policy;
- deterministic known-predicate producers can bypass dynamic growth safely.

Integration tests:

- `TriplesExtractedEvent` paths are rerouted through admission;
- direct `AddTriples` producers no longer feed raw unknown predicates into store validation;
- growth-enabled `AssertFact` paths no longer fail before schema-admission logic runs;
- one unknown candidate no longer rolls back valid admitted triples in the same batch;
- graph admission dead-letter replay is auditable across schema version change.

Regression tests:

- repeated `batch graph update error` logs are no longer the primary unknown-predicate handling path;
- graph store validator still rejects truly inconsistent admitted triples;
- deterministic seeded-predicate producers keep current behavior.

## OpenSpec Mapping

This document is a design memo, not an implementation contract by itself. Before code changes, it should be mapped into OpenSpec changes with rollback boundaries that match runtime risk.

Recommended split:

- Change A: admission boundary hardening
  - Phase A1: observe-only admission instrumentation with no write-path enforcement
  - Phase A2: enforce admission routing on dynamic producers
  - producer inventory
  - single source of truth validation
  - event and direct-store rerouting
  - batch pre-filtering while keeping atomic batch writes
- Change B: adaptive shadow growth
  - Depends on Change A
  - schema candidate persistence
  - governance-aware shadow registration path using `proposed -> shadow`
  - budget enforcement
  - retrieval provenance for shadow edges
- Change C: promotion and replay hardening
  - Depends on Change B
  - shadow usage metrics
  - manual-to-automatic promotion boundary
  - dead-letter replay idempotency
  - operator review surfaces

These should be separate changes or clearly separated waves because rollback boundaries differ. Admission boundary hardening is a safety fix. Automatic shadow growth changes runtime write behavior. Promotion automation changes governance behavior again.

## Acceptance Criteria

- The design explicitly reuses current ontology lifecycle, graph buffering, and governance primitives.
- There is one authoritative predicate validity source consumed by both `ValidateTriple` and graph store validation.
- All dynamic producer call sites are inventoried and assigned an admission migration path.
- Batch rollback is addressed by pre-admission filtering while retaining atomic admitted-batch writes.
- Shadow predicate retrieval and FTS5 semantics are defined for first wave.
- Promotion, budget, replay, and generic-predicate policy boundaries are explicit rather than implied.
- The design explicitly resolves the governance-enabled `RegisterPredicate` status-override conflict for shadow admission.
- Phase A1 observe-only behavior is unambiguous and Phase A2 is the first enforcing phase.
- After implementation, unknown-predicate graph write `ERROR` logs from normal LLM producer paths should drop by at least 95% during a representative observation window.
- After implementation, one rejected dynamic candidate must not cause loss of valid admitted triples from the same input batch.

## Growth-Enabled AssertFact Scope

Wave 1 does not assume every `AssertFact` caller participates in schema growth automatically.

Current direct caller families include:

- ontology tools and ontology actions;
- ontology P2P fact import paths;
- any future callers that submit arbitrary user- or peer-supplied predicates.

Definition:

- `growth-enabled` means a caller path is explicitly configured to invoke admission before `ValidateTriple` rejection for unknown predicates or types.
- deterministic internal callers that only assert already-known predicates are not automatically growth-enabled.
- the growth-enabled switch belongs to service or runtime configuration, not ad hoc caller behavior.
