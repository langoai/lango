# Proposal: Semantic Backend Decoupling + Legacy Path Cleanup (Steps 14-15)

## Problem

sqlite-vec is a required compile-time dependency even when users only need FTS5 text search. The `init()` function in `sqlite_vec.go` auto-loads the extension at process startup regardless of configuration. Additionally, shadow comparison infrastructure (Step 6) has served its purpose — the coordinator has been validated across Steps 6-13 — and dead code from incremental migration (assembly wrappers, `oldRetrieved` variable) clutters the codebase.

## Solution

**Step 14 — semantic-backend-decoupling:** Isolate sqlite-vec behind a `//go:build vec` tag. Introduce a `NewVectorStore` factory function that dispatches by build tag: when `vec` is present, returns `SQLiteVecStore`; when absent, returns `ErrVecNotCompiled`. Default build becomes FTS5-only. Full build uses `-tags "fts5,vec"`.

**Step 15 — legacy-path-cleanup:** Remove shadow comparison infrastructure (`shadow.go`, coordinator Shadow field/methods, shadow goroutine in GenerateContent). Promote coordinator to Phase 1 primary for factual layers (UserKnowledge, AgentLearnings, ExternalKnowledge). Reduce old retriever to non-factual layers only (RuntimeContext, ToolRegistry, SkillPatterns, PendingInquiries). Remove dead code (`assembleMemorySection`, `oldRetrieved`). Refactor run summary cache to store raw data instead of formatted string.

## Scope

- Build-tag split for sqlite-vec (embedding package)
- Factory function pattern for VectorStore creation
- Wiring graceful degradation for vec-less builds
- Test gating with `//go:build vec`
- Build infrastructure (Makefile, Dockerfile) updates
- Documentation (README, build-test docs)
- Shadow infrastructure removal (retrieval package)
- Coordinator promotion to Phase 1 primary
- Dead code removal (assembly wrappers, shadow comparison)
- Run summary cache refactoring
- Config cleanup (remove Shadow from RetrievalConfig)

## Non-Goals

- LIKE fallback removal (remains as FTS5 safety net)
- Auto-enable logic changes (Step 8 scope preserved)
- go.mod dependency removal (sqlite-vec stays in module graph)
