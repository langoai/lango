## Context

Steps 1-4 complete. Knowledge has 6 categories but only 2 are actively used by learning analyzers. Pattern/correction types are routed to Learning only. No temporal classification exists. Duplicate mappers across packages.

## Goals / Non-Goals

**Goals:**
- Full 6-category taxonomy used by ALL analyzers consistently
- Temporal classification (evergreen vs current_state) as tags
- Dual-save: pattern/correction → Knowledge + Learning
- Content-dedup to prevent version churn (Step 4 interaction)
- Shared mapper as single source of truth
- Learning mapper error return (no silent fallback)

**Non-Goals:**
- New category enum values (6 categories are sufficient)
- Category-aware retrieval filtering (later step)
- Learning temporal classification (separate step)
- Taxonomy changes to learning categories

## Decisions

### D1. Dual-save for pattern/correction across all paths
**Choice**: When ANY ingestion path encounters type `pattern` or `correction`, save as Knowledge AND as Learning. Learning entry is for backward-compat error-pattern matching.

**Rationale**: Pattern and correction are valid knowledge categories. Previously they were only saved as Learning, losing them from knowledge retrieval.

### D2. Content-dedup: (category, content) equality = no-op
**Choice**: In `saveKnowledgeOnce`, if latest version has same `(category, content)`, return nil without creating a new version. `source`, `tags`, and temporal hint changes alone do NOT justify a new version.

**Rationale**: Multiple analyzers may re-extract the same fact from different turns. Without dedup, Step 4's versioning creates meaningless v2/v3/v4 with identical content.

### D3. Temporal as tag, not field
**Choice**: Temporal classification stored as `"temporal:evergreen"` or `"temporal:current_state"` in KnowledgeEntry.Tags, not as a new schema field.

**Rationale**: Avoids Ent schema migration. Tags are already used for metadata. Temporal is a classification hint, not a structural property.

### D4. Shared saveAnalysisResult helper
**Choice**: ConversationAnalyzer.saveResult() and SessionLearner.saveSessionResult() delegate to a shared `saveAnalysisResult()` in parse.go, parameterized by `saveResultParams` (key prefix, trigger prefix, source label).

**Rationale**: The two methods were nearly identical (~40 lines each). Consolidation eliminates duplication while preserving behavioral compatibility.

### D5. EventBus pattern for graph triples
**Choice**: All graph triple emission uses `eventbus.Bus.Publish(TriplesExtractedEvent)` pattern, not legacy `GraphCallback`.

**Rationale**: Current branch (feature/memory-renewal) standardized on eventbus. Workers initially used main-branch GraphCallback pattern; corrected during integration.

## Risks / Trade-offs

**[Risk] LLM prompt changes affect extraction quality** → Mitigated by additive change (more types requested, same format). LLM may produce new types it didn't before, but all are valid categories.

**[Trade-off] Dual-save increases write volume** → Only for pattern/correction types, which are a small fraction. Learning store is already high-volume.

**[Trade-off] Content-dedup comparison is string-equality** → Minor whitespace differences create new versions. Acceptable — LLM extraction is non-deterministic but typically produces identical strings for same facts.
