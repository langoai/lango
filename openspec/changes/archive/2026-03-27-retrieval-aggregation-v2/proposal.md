# Proposal: retrieval-aggregation-v2

## Problem
The coordinator resolves duplicate (Layer, Key) findings with a naive "highest Score wins" rule. An auto-extracted observation with high FTS5 relevance can beat a user correction with lower relevance. There is no awareness of knowledge authorship, version chain, or recency.

## Solution
Replace score-only dedup with evidence-based merge using a deterministic priority chain: **authority → version (supersedes) → recency → score**. Enrich Finding struct with provenance metadata (Source, Tags, Version, UpdatedAt) and update agents to populate it. Fix save_knowledge tool to set default Source="knowledge" for user-explicit saves.

## Scope
- Finding provenance enrichment (4 new fields)
- FactSearchAgent + TemporalSearchAgent populate provenance
- ContextSearchAgent: no change (RAGResult lacks provenance, falls through to Score)
- mergeFindings replaces dedupFindings with authority-first resolution
- sourceAuthority ranking map (6 known sources)
- save_knowledge default Source fix
- No config changes, no backfill of existing data
