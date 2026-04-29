# Retrieval Orchestration Efficiency Audit Design

## Purpose

This workstream audits the efficiency boundary where retrieval, context assembly, and multi-agent orchestration meet. The goal is to produce a code-grounded data-flow map and land only low-risk quick wins that reduce avoidable repeated work.

This is not a broad runtime rewrite. It is a map-guided optimization pass.

## Scope

Included surfaces:

- `internal/knowledge` context retrieval, prompt assembly, token truncation, and store search paths
- `internal/retrieval` agentic retrieval coordination, merge, sort, and truncation paths
- Graph RAG entry points where content retrieval expands into graph traversal
- `internal/agentrt` coordinating executor paths that may trigger repeated model or delegation work
- `internal/orchestration` sub-agent routing and dynamic agent table construction
- `internal/streamx` fan-in helpers used for child-agent output merging

Excluded surfaces:

- Ranking policy changes for retrieval results
- Graph RAG algorithm replacement
- P2P protocol semantics changes
- Large agent runtime redesign
- Configuration model redesign
- Public documentation expansion unless implementation changes user-facing behavior

## Design Direction

Use a Map-Guided Quick Wins approach:

1. Build a shallow architecture map from the real code paths.
2. Identify repeated retrieval, merge, sort, token-estimation, allocation, or fan-in bookkeeping costs.
3. Select only one or two quick wins that are local, safe, and testable.
4. Leave broader findings as follow-up candidates instead of folding them into this change.

## Data Flow To Audit

The first audit pass should trace a user turn through:

1. Request entry through TurnRunner / ADK model adapter.
2. `ContextAwareModelAdapter` context injection.
3. `ContextRetriever` layer retrieval and prompt assembly.
4. `RetrievalCoordinator` parallel retrieval agents and merge path.
5. Graph RAG retrieval and expansion, when configured.
6. Observational memory retrieval.
7. Orchestrator routing and sub-agent delegation.
8. Stream fan-in for child agent output.

The map should explicitly mark:

- where the same query can be searched more than once in one turn
- where retrieval results are sorted, deduplicated, or token-estimated multiple times
- where slice/map allocation can be pre-sized without changing behavior
- where orchestration creates repeated context or routing work
- where fan-in emits progress or wraps streams with avoidable bookkeeping

## Quick Win Selection Criteria

A quick win must satisfy all of these:

- Local: the change stays inside one or two packages.
- Safe: retrieval result ranking, evidence priority, delegation behavior, and P2P semantics remain unchanged.
- Testable: an existing unit test can be extended, or a small benchmark/unit test can verify the behavior.
- Visible: the inefficiency is clear from the code path, not speculative.

Candidate classes:

- repeated token estimation in truncation paths
- unnecessary sorting or duplicate merge work
- avoidable slice/map growth in retrieval and orchestration paths
- repeated lowercasing or keyword transformation in local search loops
- fan-in wrapper work that can be simplified without altering event semantics

## Testing Strategy

Testing should preserve behavior first and measure where practical:

- Extend existing retrieval tests for unchanged merge priority and truncation behavior.
- Add focused tests for any selected allocation or caching behavior.
- Add a small benchmark only when the improvement is performance-specific and benchmark setup is stable.
- Keep Graph RAG, orchestration, and fan-in semantics unchanged.

Full verification after implementation:

- `go build ./...`
- `go test ./...`
- OpenSpec workflow validation for the implemented change

## OpenSpec Plan

Implementation should create a dedicated OpenSpec change named `retrieval-orchestration-efficiency-audit`.

Expected artifacts:

- proposal describing the map-guided efficiency audit
- specs covering behavior-preserving efficiency improvements
- design referencing this internal design document
- tasks splitting audit map, quick-win implementation, tests, and downstream docs checks

After implementation, run the repository-required OpenSpec flow:

- ff
- apply
- verify
- sync
- archive

## Completion Criteria

The workstream is complete when:

- a code-grounded retrieval + orchestration flow map exists in the implementation artifacts
- one or two safe quick wins are implemented, or the audit explicitly records why no quick win met the criteria
- behavior-preserving tests cover the changed paths
- no public docs are changed unless CLI/TUI/user-facing behavior changes
- `go build ./...` and `go test ./...` pass
- the OpenSpec change is verified, synced, and archived

## Follow-Up Candidates

Potential follow-up tracks, if supported by the audit:

- deeper retrieval benchmarking and latency tracing
- Graph RAG traversal limit tuning
- cross-turn retrieval cache design
- orchestration context-injection deduplication
- agent fan-in throughput and cancellation profiling
