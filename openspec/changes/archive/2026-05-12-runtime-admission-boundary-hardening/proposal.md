# Runtime Admission Boundary Hardening

## Why

Dynamic runtime graph inputs currently surface unknown-predicate failures only after graph-store validation runs. Before changing write behavior, the runtime needs an observe-only admission boundary that classifies the supported event-bus producer sources and the `content.saved` extraction path without changing current write routing.

## What Changes

- Add an observe-only graph admission policy for the supported runtime graph inputs in this slice.
- Publish graph admission telemetry, extractor dropped-unknown baselines, graph write-failure baselines, `unmapped-source` telemetry, event-bus producer-source / producer-group attribution, the synthetic `content_saved_extractor` source label, and validator-source tags to observability and cockpit status.
- Reuse one ontology predicate validator closure as the primary predicate-validity source across admission classification and graph-store validation when ontology is available. If ontology is unavailable, graph-store validation falls back to the built-in hardcoded graph predicate validator and observe-only admission degrades to an `unvalidated` observation mode.
- Fix the observe-only admission decision taxonomy at the batch level and record triple-level counts for `known`, `unknown`, and `unvalidated` predicates.

## Terminology

- **Producer source**: the stable runtime label taken from `TriplesExtractedEvent.Source`.
- **Telemetry source label**: a stable non-event-bus telemetry label. In this slice the only synthetic label is `content_saved_extractor`.
- **Producer group**: the fallback-confidence configuration group used for event-bus producer sources. In this slice the only groups are `learning` and `librarian`.
- **Validator-source tag**: the stable telemetry tag naming the predicate validator source used for observe-only classification.
- The stable ontology-backed validator-source value in this slice is `ontology_registry`.
- **Unmapped source**: a raw `TriplesExtractedEvent.Source` label that is outside the supported event-bus producer-source set in this slice and therefore is not assigned to one of the known producer groups.
- **Admission decision taxonomy**: `known`, `unknown`, and `unvalidated` triple counts computed for each observed batch. `unvalidated` is used only when validator-based classification is unavailable.
- **Unavailable validator source**: the stable `validator-source` value used when validator-based classification is unavailable and the batch is observed as fully `unvalidated`.

## Out Of Scope

- write filtering or dropping
- CLI import
- `AssertFact`/ontology fact assertion paths
- adaptive shadow growth
