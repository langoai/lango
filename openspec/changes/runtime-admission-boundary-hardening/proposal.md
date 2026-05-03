# Runtime Admission Boundary Hardening

## Why

Dynamic runtime graph producers currently produce unknown-predicate failures that surface too late at graph validation time. Before changing write behavior, the runtime needs an observe-only admission layer to classify those batches and measure where the failures come from.

## What Changes

- Add an observe-only graph admission policy for runtime app producers.
- Publish graph admission telemetry, dropped-unknown extractor telemetry, and graph write failure baselines to observability and cockpit status.
- Share one ontology predicate validator closure across graph admission and graph-store validation.

## Out Of Scope

- write filtering or dropping
- CLI import
- `AssertFact`/ontology fact assertion paths
- adaptive shadow growth
