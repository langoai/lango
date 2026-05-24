## Summary

Remove the production `panic` path from ontology schema digest computation while preserving the existing `ComputeDigest` string API and successful export behavior.

## Motivation

Ontology schema export is reachable through local service calls and the P2P ontology bridge. Digest computation currently uses a helper that panics if canonical JSON marshaling fails. Even though the current slim schema types are JSON-safe, production export paths should return errors instead of crashing the process.

## Scope

- Replace panic-based digest marshaling with checked digest computation.
- Keep `ComputeDigest(types, predicates) string` behavior for existing callers.
- Propagate digest computation failures through `ExportSchema` as returned errors.
- Add an executable regression guard preventing production `panic` calls in `internal/ontology`.

## Non-Goals

- No changes to schema bundle JSON format.
- No changes to import conflict semantics or governance behavior.
- No changes to P2P ontology bridge request/response contracts.
