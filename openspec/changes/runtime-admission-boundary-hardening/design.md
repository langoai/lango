# Design

This change implements only `Change A / Phase A1`.

The runtime computes admission decisions for:
- `TriplesExtractedEvent` batches
- app-path extracted triples only where the current runtime already surfaces them without widening extractor behavior
- extractor-local dropped-unknown events for the current `content.saved` extraction path

In all cases, current write routing remains unchanged. The policy is observe-only: it records what would be admitted or rejected, emits telemetry, and then forwards the **original triple slice** through the existing enqueue/store path unchanged.
