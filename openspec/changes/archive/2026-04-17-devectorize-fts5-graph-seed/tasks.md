## 1. Graph And Context Wiring

- [x] 1.1 Replace the vector-specific GraphRAG seed interface with a generic content retriever contract.
- [x] 1.2 Rewire GraphRAG phase 1 to use hydrated knowledge FTS5 results from app wiring.
- [x] 1.3 Remove the standalone vector retrieval path from the ADK context-aware adapter.

## 2. Retrieval And Tool Surface

- [x] 2.1 Remove the vector-backed context-search retrieval agent from the coordinator wiring.
- [x] 2.2 Remove the `rag_retrieve` tool and related saveable/profile metadata.
- [x] 2.3 Update prompt/event wording from RAG-specific language to retrieved-context language where the runtime path changed.

## 3. Verification

- [x] 3.1 Update parity/regression tests for the removed RAG tool and coordinator changes.
- [x] 3.2 Verify with `go build -tags fts5 ./...`.
- [x] 3.3 Verify with `go test -tags fts5 ./...`.
