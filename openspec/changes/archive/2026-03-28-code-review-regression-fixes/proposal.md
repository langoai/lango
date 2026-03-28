# Proposal: Code Review Regression Fixes

## Problem

Code review of Steps 14-15 (semantic-backend-decoupling + legacy-path-cleanup) found 4 merge blockers: 2 regressions from the current branch and 2 pre-existing contract violations. `go test ./...` fails on `checks_test.go` and `forms_impl_test.go`.

## Fixes

1. **[P1] DefaultConfig() defaults regression**: `dev` branch had RunLedger, Provenance, Sandbox defaults that were lost in the current branch. Restored exact values from `origin/dev:internal/config/loader.go:142-196`.

2. **[P1] Settings TUI explicitKeys regression**: `settings.go` saves with `nil` explicitKeys, causing auto-enable to override user-disabled context features on restart. Fixed by marking all context-related keys as explicit on settings save. Exported `ContextRelatedKeys()` helper (returns slice copy).

3. **[P2] RAG enabled flag ignored**: `wiring_embedding.go` always created `ragService` and `wiring_knowledge.go` always registered `ContextSearchAgent` regardless of `embedding.rag.enabled`. Gated both on the flag.

4. **[P2] Relevance score clamping**: `BoostRelevanceScore` could exceed maxScore, `DecayAllRelevanceScores` left a gap above minScore. Implemented two-step clamping (add/cap for boost, subtract/floor for decay) with proper error handling.
