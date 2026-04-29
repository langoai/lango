# Proposal: Retrieval Orchestration Efficiency Audit

## Summary

Audit the boundary where context retrieval, agentic retrieval, Graph RAG, and multi-agent orchestration meet. Land only behavior-preserving quick wins that reduce avoidable repeated work.

## Motivation

Recent feature work expanded retrieval, memory, Graph RAG, structured agent runtime, and multi-agent orchestration. The code now has several paths where a user turn can trigger retrieval, merge, sort, token estimation, prompt assembly, and delegation bookkeeping. Before larger performance work, the system needs a code-grounded flow map and a small set of local optimizations.

## Non-Goals

- Change retrieval ranking or evidence priority.
- Replace Graph RAG traversal algorithms.
- Redesign agent runtime or P2P semantics.
- Redesign configuration.
- Expand public docs unless behavior becomes user-facing.

## Success Criteria

- A code-grounded retrieval/orchestration flow map exists in `design.md`.
- One or two local quick wins are implemented or explicitly rejected with evidence.
- Retrieval and context truncation behavior remains unchanged.
- `go build ./...` and `go test ./...` pass.
