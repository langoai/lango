# Design: Retrieval Orchestration Efficiency Audit

## Code-Grounded Flow Map

1. User turn enters TurnRunner / ADK model adapter.
2. `ContextAwareModelAdapter` injects retrieved context before model execution.
3. `ContextRetriever.Retrieve` extracts keywords, walks requested layers, and calls knowledge, skills, external refs, learnings, tool registry, runtime context, and pending inquiries as configured.
4. Agentic retrieval uses `RetrievalCoordinator.Retrieve`, runs retrieval agents in parallel, appends all findings, merges by `(Layer, Key)`, sorts by score, and truncates by token budget.
5. Graph RAG runs a content retrieval phase and expands graph nodes when configured.
6. Observational memory can add observations and reflections to prompt context.
7. Structured orchestration routes tool-requiring work through `BuildAgentTree`, sub-agent routing entries, and `CoordinatingExecutor`.
8. Child agent output can be merged through `streamx.AgentStreamFanIn`.

## Quick Wins Selected

### Retrieval aggregation preallocation

`RetrievalCoordinator.Retrieve` already knows the number of retrieval-agent result buckets. After agent completion it can count total findings once and allocate `allFindings` with exact capacity. `mergeFindings` can keep its existing map capacity and return slice allocation. This preserves ordering because final ordering still comes from `sortFindingsByScore`.

### Context truncation token-cost cache

`knowledge.TruncateResult` currently estimates all item token costs to check whether truncation is needed, then estimates retained candidates again while rebuilding the truncated result. Replace the second estimation with a per-item token-cost list built during the first pass. This preserves layer priority and item ordering.

## Follow-Up Candidates

- Cross-turn retrieval cache design.
- Graph RAG traversal benchmark and limit tuning.
- Orchestration context-injection deduplication.
- Agent fan-in cancellation and throughput profiling.
