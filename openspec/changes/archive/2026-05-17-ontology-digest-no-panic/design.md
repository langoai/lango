## Context

`internal/ontology/exchange.go` computes schema bundle digests from sorted slim type and predicate arrays. The exported `ComputeDigest` function currently returns a string and is covered by determinism tests. The production export path calls it from `exportSchema`.

## Design

Introduce an unexported digest helper that returns `(string, error)`. It will sort the digest payload, JSON-marshal it, and return a wrapped error if marshaling fails. `exportSchema` will call this checked helper and return an actionable `compute schema digest` error on failure.

Keep the exported `ComputeDigest` signature as a compatibility wrapper around the checked helper. If the impossible marshal error occurs through this compatibility API, return an empty digest rather than panic. Existing deterministic inputs continue producing the same hash.

## Testing

Add an AST-based ontology package guard that rejects `panic` calls in non-test Go files. Existing digest and export tests continue to verify stable digest output and successful schema bundle generation.

## Risks

The main risk is changing digest bytes or ordering. Existing `ComputeDigest` order-independence and `ExportSchema` digest stability tests cover the successful path, while the new guard prevents reintroducing panic-based failure handling.
