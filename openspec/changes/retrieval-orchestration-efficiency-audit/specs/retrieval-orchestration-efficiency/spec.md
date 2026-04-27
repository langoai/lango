## Purpose

Behavior-preserving efficiency requirements for the retrieval and orchestration boundary.

## Requirements

### Requirement: Retrieval orchestration flow map
The change SHALL document the real code path from a user turn through context retrieval, agentic retrieval, Graph RAG, memory retrieval, orchestration delegation, and stream fan-in.

#### Scenario: Flow map identifies repeated work candidates
- **WHEN** the design artifact is reviewed
- **THEN** it lists where same-turn repeated query/search, merge, sort, token estimation, allocation, or fan-in bookkeeping can occur

### Requirement: Retrieval aggregation preallocation preserves behavior
`RetrievalCoordinator.Retrieve` and `mergeFindings` SHALL pre-size aggregation containers where cardinality is known without changing deduplication, score ordering, authority priority, or token truncation behavior.

#### Scenario: Authority merge still wins over score
- **WHEN** same-layer findings share a key but have different source authority
- **THEN** the higher-authority finding is retained even if its score is lower

#### Scenario: Different layers still preserve same key
- **WHEN** findings share a key across different context layers
- **THEN** both findings remain present after merge

### Requirement: Context truncation estimates each item once
`knowledge.TruncateResult` SHALL avoid estimating token count for the same item more than once in a single truncation call.

#### Scenario: Existing truncation behavior is preserved
- **WHEN** a result exceeds the budget
- **THEN** items are retained in the existing layer priority order until the budget is exhausted

#### Scenario: Too-small budget still returns zero items
- **WHEN** the first item exceeds the budget
- **THEN** the truncated result contains zero items
