# Design

This change implements only `Change A / Phase A1`.

Producer terminology in this slice is fixed as follows:
- **Producer source** = the stable runtime source label taken from `TriplesExtractedEvent.Source`.
- **Telemetry source label** = a stable synthetic label for a non-event-bus observed path.
- **Producer group** = the fallback-confidence configuration group for event-bus producer sources.
- `conversation_analysis`, `session_learning`, and `learning` map to the `learning` producer group.
- `proactive_librarian` maps to the `librarian` producer group.
- `content_saved_extractor` is the only synthetic telemetry source label in this slice and is used on observe-only telemetry emitted for returned triples and dropped-unknown baselines produced by the `content.saved` extraction path.
- Any other raw `TriplesExtractedEvent.Source` value remains visible as an `unmapped-source` telemetry signal and still follows the same graph write operation without observe-only admission classification.

The runtime computes observe-only admission decisions for:
- supported event-bus batches whose `TriplesExtractedEvent.Source` is one of `conversation_analysis`, `session_learning`, `learning`, or `proactive_librarian`
- triples returned from the `content.saved` extraction path; observe-only telemetry for that path SHALL use the synthetic `content_saved_extractor` source label and SHALL publish those telemetry events on the runtime event bus

Separately, this slice records a pre-admission extractor baseline for dropped-unknown events emitted by the `content.saved` extraction path before graph admission runs and tagged with the stable `content_saved_extractor` telemetry source label.

Each observe-only admission decision is batch-scoped. For every observed triple slice, the runtime computes:
- `batch_count = 1` for the observed slice
- `known_count` = number of triples whose predicate is accepted by validator-based classification
- `unknown_count` = number of triples whose predicate is rejected as unknown by validator-based classification
- `unvalidated_count` = number of triples left unclassified because validator-based classification is unavailable

`known_count + unknown_count + unvalidated_count` SHALL equal the number of triples in the observed slice.

When validator-based classification is unavailable, the runtime still emits one observe-only admission observation for the batch with:
- `known_count = 0`
- `unknown_count = 0`
- `unvalidated_count = len(observed slice)`
- `validator_source = "unavailable"`

When validator-based classification is available through the ontology service closure, the runtime uses the stable validator-source value `ontology_registry`.

This slice does not introduce new producer families or new source adapters. Runtime admission config is limited to mode selection (`off` or `observe`) plus fallback confidence defaults for the learning producer group and the librarian producer group. These settings are stored under the existing `ontology.governance.*` namespace for config compatibility, but they always remain directly visible and editable on the runtime admission settings surface rather than inheriting governance-enabled gating semantics.

In all cases, observe-only mode MUST NOT drop, rewrite, or reroute the original triple slice. The policy classifies that **original triple slice** against the shared predicate-validity source, emits event-bus admission telemetry tagged with stable producer-source values plus producer-group identifiers and validator-source tags, emits `content.saved` admission telemetry tagged with the synthetic `content_saved_extractor` telemetry source label and validator-source tag, preserves unknown event-bus source labels as `unmapped-source` signals instead of collapsing them, and then leaves the original triple slice to proceed through the same graph write operation it would have used without observe-only admission. Extractor dropped-unknown baselines and aggregate graph write-failure baselines remain separate telemetry families rather than admission decisions.
