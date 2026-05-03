# Adaptive Shadow Ontology Growth Design

## Context

Lango's knowledge, learning, ontology, graph, and FTS5 systems are intended to support an agent that can collect, classify, organize, retrieve, and evolve knowledge over time. The current graph ingestion path treats unknown predicates as storage errors. This makes the ontology act like a static allowlist instead of a self-improving schema layer.

The observed failure mode is a batch graph update error such as `unknown predicate "includes"` or `unknown predicate "defines"`. These predicates are semantically useful discoveries, but they are not part of the seeded ontology predicate set. Because graph writes are batched, one unknown predicate can reject an entire graph update batch.

## Problem

Unknown ontology terms are currently handled too late in the pipeline. They reach the graph store validator before they have been normalized, proposed, shadowed, or quarantined.

This causes three product-level failures:

- User-facing agent behavior becomes coupled to background knowledge growth failures.
- LLM-discovered schema extensions are treated as errors instead of growth signals.
- Valid triples in the same batch can be lost because one invalid predicate aborts the batch transaction.

## Goals

- Preserve user response continuity even when ontology or graph growth fails.
- Convert unknown predicates and types into ontology growth candidates.
- Allow high-confidence schema discoveries to become usable through `shadow` status.
- Normalize obvious semantic aliases such as `includes` to canonical predicates such as `contains`.
- Keep graph storage validation as a final integrity guard, not the first discovery mechanism.
- Provide observable metrics and operator review surfaces for automatic schema growth.

## Non-Goals

- Do not make every LLM-created relation immediately active.
- Do not remove graph predicate validation.
- Do not bypass ontology governance or schema lifecycle rules.
- Do not make public documentation claims before the feature is implemented and wired.

## Recommended Policy

Use an `adaptive-shadow` policy.

Unknown ontology terms move through a deterministic admission process:

1. Check whether the predicate or type is already known.
2. Try semantic normalization against existing canonical schema entries.
3. Score confidence from source, repetition, subject/object type evidence, and lexical similarity.
4. Check governance rate limits and schema growth budgets.
5. Admit, map, shadow, propose, quarantine, or dead-letter the candidate.

The default decisions are:

- `known`: Store the triple as-is.
- `mapped`: Rewrite to a canonical predicate and store the mapped triple.
- `shadow`: Register a high-confidence new predicate or type as `shadow`, refresh validation, and store the triple.
- `proposed`: Persist the schema candidate and sample evidence, but do not store the graph triple.
- `quarantined`: Persist evidence for suspicious, generic, conflicting, or explosive schema candidates.
- `dead-lettered`: Persist failed admission or storage attempts for later replay.

`shadow` means "usable but experimental." Shadow predicates may participate in graph storage and retrieval, but they are reported separately from active schema and require evidence before promotion.

## Architecture

Introduce an ontology admission layer between all raw triple producers and the graph buffer.

```text
Knowledge / Learning / Librarian / Extractor
  -> raw TripleCandidate
  -> TripleAdmissionPolicy
  -> OntologyGrowthEngine
  -> AdmissionResult
  -> accepted triples only
  -> GraphBuffer
  -> GraphStore
```

### Components

`TripleCandidate`

Represents a raw triple emitted by an LLM, learning engine, librarian process, or extractor. It includes subject, predicate, object, optional types, source, confidence, session key, turn ID, and producer name.

`TripleAdmissionPolicy`

Provides the single entry point for graph admission. It accepts a `TripleCandidate` and returns an `AdmissionResult`. It is responsible for normalizing, filtering, and routing candidates before they reach `GraphBuffer`.

`OntologyGrowthEngine`

Evaluates unknown predicates and types. It checks existing ontology state, alias rules, semantic similarity, confidence, governance policy, and schema growth budget. It may create proposed or shadow schema entries through `OntologyService`.

`SchemaProposalStore`

Persists discovered candidates, source evidence, sample triples, status, counts, last seen time, last error, and promotion hints. This store is the review and replay substrate.

`GraphAdmissionDeadLetterStore`

Records admission or graph write failures that should not interrupt user response flow. Dead letters include enough data to replay the candidate after schema repair.

## Data Flow

Known predicate path:

```text
raw triple: user_preference related_to coding_style
  -> validator recognizes related_to
  -> accepted
  -> GraphBuffer
```

Alias mapping path:

```text
raw triple: project includes module
  -> includes is unknown
  -> semantic alias maps includes -> contains
  -> mapped triple: project contains module
  -> GraphBuffer
```

Adaptive shadow path:

```text
raw triple: term defines concept
  -> defines is unknown
  -> high confidence, repeated, not a generic predicate
  -> register PredicateDefinition{Name: "defines", Status: shadow}
  -> refresh predicate validator
  -> GraphBuffer
```

Low-confidence path:

```text
raw triple: thing uses stuff
  -> uses is unknown and too generic
  -> proposed or quarantined
  -> graph storage skipped
  -> user response continues
```

Storage failure path:

```text
admitted triple
  -> GraphBuffer write fails
  -> retry per triple or dead-letter
  -> valid independent triples are not lost silently
```

## Error Handling

Ontology growth failures are not user-response failures. They are contained as knowledge-write outcomes.

- Unknown predicates should not be logged as graph update `ERROR` unless they reach the graph store after admission.
- Candidate creation, mapping, proposal, and quarantine should be logged as structured schema events.
- Validator/cache inconsistency after shadow registration is an `ERROR`.
- Batch graph failures should include source counts and should attempt triple-level isolation or dead-lettering.
- All dead-letter records must be replayable after schema repair.

## Governance And Safety

Automatic growth requires hard budgets.

- Daily auto-shadow budget limits how many new predicates or types can become usable without review.
- Monthly schema explosion budget limits total new schema growth.
- Generic predicates such as `has`, `is`, `uses`, `related`, `thing`, and `misc` are quarantined by default unless explicit mapping rules exist.
- Shadow entries must track source evidence and usage before active promotion.
- Active promotion remains a governance transition, not an ingestion side effect.

## Observability

The system should expose these counters and status fields:

- Unknown predicate discoveries.
- Unknown type discoveries.
- Canonical mapping count.
- Auto-shadow creation count.
- Proposed, quarantined, and dead-lettered candidate counts.
- Top unknown predicates by frequency.
- Shadow predicate usage count.
- Graph admission rejection reasons.
- Schema budget remaining.
- Validator/cache mismatch count.

These metrics should be available to internal status surfaces before public documentation is expanded.

## Testing Strategy

Unit tests:

- `includes` maps to `contains`.
- `defines` can be auto-registered as `shadow` when confidence and budget allow.
- Low-confidence unknown predicates become proposed-only.
- Generic predicates are quarantined.
- Existing active and shadow predicates pass admission unchanged.
- Governance budget exhaustion prevents auto-shadow creation.

Integration tests:

- A `TriplesExtractedEvent` with an unknown predicate does not cause user-response failure.
- A batch with one invalid candidate still stores valid admitted triples.
- Shadow registration refreshes `PredicateValidator()` immediately.
- Graph store validator remains a final guard for admitted triples.
- Dead-letter replay succeeds after the relevant predicate is promoted or mapped.

Regression tests:

- Unknown predicates no longer produce repeated `batch graph update error` logs as the primary handling path.
- Valid triples are not discarded because another candidate in the same batch is unknown.

## Rollout

1. Add the admission layer in observe-only mode. It records decisions but does not change graph routing.
2. Enable mapping and proposed-only handling for unknown predicates.
3. Enable bounded auto-shadow for high-confidence candidates.
4. Add operator status and replay surfaces.
5. Promote stable shadow entries through existing governance workflows.

## Acceptance Criteria

- Unknown LLM-discovered predicates are converted into mapped, proposed, shadow, quarantined, or dead-lettered outcomes.
- User responses continue even when ontology growth or graph storage fails.
- `GraphBuffer` no longer receives raw unknown predicates from normal LLM extraction paths.
- High-confidence shadow predicate registration refreshes graph validation immediately.
- Batch graph writes no longer lose all valid triples because one candidate is unknown.
- Operators can inspect top schema candidates, shadow usage, and dead-lettered graph admissions.
