# Proposal: temporal-agent-v1

## Problem
The retrieval coordinator has FactSearchAgent (keyword) and ContextSearchAgent (vector) but neither uses temporal metadata. Queries about "what changed recently" or "when was X updated" get no special treatment — results are ranked by keyword relevance or vector similarity, not freshness.

## Solution
Add a TemporalSearchAgent that leverages the version chain (Step 4) and `updated_at` timestamps to surface freshness-relevant results. Score by recency (0-1 range) and enrich content with version/time metadata.

## Scope
- v1: LayerUserKnowledge only (learnings lack version chains)
- Recency scoring with 1-week decay window
- Content enrichment with `[vN | updated Xh ago]` prefix
- Always registered in coordinator (no optional dependencies)
