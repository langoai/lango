## No Spec Changes

This change is an internal quality improvement. All fixes preserve existing API contracts and test behavior. No new capabilities are introduced and no existing spec requirements are modified.

Affected specs (verified unchanged):
- `ontology-governance`: `PromoteType/PromotePredicate` signature preserved (reason parameter kept)
- `ontology-registry`: Registry interface unchanged
- `ontology-actions`: ActionExecutor contract unchanged (logging is observability-only)
- `truth-maintenance`: TruthMaintainer interface unchanged
- `entity-resolution`: EntityResolver interface and MergeResult struct unchanged
- `property-store`: PropertyStore public API extended (batch method added, existing methods unchanged)
