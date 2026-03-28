# Design: retrieval-aggregation-v2

## Architecture

The coordinator's pipeline changes from score-only dedup to evidence-based merge:

```
Before: agents → dedupFindings (highest Score wins) → sort → truncate
After:  agents → mergeFindings (authority→version→recency→score) → sort → truncate
```

## Key Decisions

### Evidence Merge Priority Chain
For same (Layer, Key) from multiple agents:
1. Authority: sourceAuthority[Source] (user-explicit=4, session_learning=3, proactive_librarian=2, others=1, unknown=0)
2. Version: higher supersedes lower (version chain IS supersedes)
3. Recency: more recent UpdatedAt wins
4. Score: search relevance as final tiebreaker

### Merge Resolution vs Global Ranking (separate concerns)
- **Merge** determines which variant of the SAME key survives (authority-first)
- **Global ranking** orders ALL surviving findings by Score
- A user correction with Score=0.3 beats auto-analysis with Score=5.0 in merge, but ranks by Score=0.3 globally

### Finding Provenance Fields
Source, Tags, Version, UpdatedAt added to Finding. Zero values = no provenance (backward compatible).

### Agent Provenance Population
- FactSearchAgent: full provenance from ScoredKnowledgeEntry.Entry
- TemporalSearchAgent: full provenance from KnowledgeEntry
- ContextSearchAgent: no provenance (RAGResult) → falls through to Score

### save_knowledge Default Source
Changed from "" to "knowledge" so user-explicit saves rank highest in authority. No backfill of existing data.

### Temporal Content Gap (known, deferred)
When FactSearch wins merge over TemporalSearch, the [vN | updated Xh ago] prefix is lost. Finding.Version and Finding.UpdatedAt carry the metadata — presentation-layer enrichment deferred to Step 12/13.
