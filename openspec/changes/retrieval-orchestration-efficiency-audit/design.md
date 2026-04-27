# Design: Retrieval Orchestration Efficiency Audit

## Code-Grounded Flow Map

1. User turn context injection starts in `ContextAwareModelAdapter.GenerateContent`.
2. `GenerateContent` extracts the last user message, resolves session/runtime state, and runs retrieval section collection in parallel, including retriever, coordinator, Graph RAG, memory, recall, and run summary data sources when configured.
3. In the current user-turn path, `ContextRetriever.Retrieve` is called with configured non-factual/context layers: runtime context, tool registry, skill patterns, and pending inquiries. The retriever still has broader layer support, but those factual store layers are not the primary path here.
4. Agentic factual retrieval uses `RetrievalCoordinator.Retrieve` when configured, covering user knowledge, agent learnings, and external knowledge through retrieval agents.
5. `GenerateContent` merges non-factual retriever results and factual coordinator results, measures section token costs including run summaries, reallocates budgets, truncates knowledge context, and assembles prompt sections.
6. Graph RAG runs a content retrieval phase and expands graph nodes when configured.
7. Observational memory can add observations and reflections to prompt context.
8. Run summaries from the active run ledger can be retrieved with `retrieveRunSummaryData` and formatted into the prompt context with `formatRunSummarySection`.
9. Structured orchestration routes tool-requiring work through `BuildAgentTree`, sub-agent routing entries, and `CoordinatingExecutor`.
10. Child agent output can be merged through `streamx.AgentStreamFanIn`.

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
