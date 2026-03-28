# Design: Code Review Regression Fixes

## D1: DefaultConfig() restore
Restore RunLedger, Provenance, Sandbox blocks from `dev` branch between Workflow and ObservationalMemory sections. Exact values from `origin/dev` — not new design.

## D2: Settings explicitKeys — "all explicit" strategy
Export `ContextRelatedKeys()` from `config` package (returns slice copy for safety). In `settings.go`, build `explicitKeys` map marking all context-related keys as true before calling `Save()`. This prevents auto-enable from overriding values the user has seen and accepted in the TUI.

**Tradeoff**: Once settings are saved, all context-related values become "user intent" and auto-enable behaves more conservatively. Appropriate for the current goal.

## D3: RAG enabled flag enforcement
Gate `RAGService` creation in `wiring_embedding.go` on `emb.RAG.Enabled`. Gate `ContextSearchAgent` registration in `wiring_knowledge.go` on `cfg.Embedding.RAG.Enabled`. Buffer is still created (needed for async embedding regardless of RAG).

## D4: Two-step relevance score clamping
**Boost**: Step 1 adds delta to rows where `score <= maxScore - delta` (safe). Step 2 caps rows where `maxScore - delta < score < maxScore` to `maxScore`.
**Decay**: Step 1 subtracts delta from rows where `score >= minScore + delta` (safe). Step 2 floors rows where `minScore < score < minScore + delta` to `minScore`.
All DB writes return errors. No transaction needed (each step is a single atomic UPDATE).

## Files Changed

| File | Change |
|------|--------|
| `internal/config/loader.go` | Restore RunLedger/Provenance/Sandbox defaults |
| `internal/config/auto_enable.go` | Export `ContextRelatedKeys()` |
| `internal/cli/settings/settings.go` | Pass all-explicit keys to Save() |
| `internal/app/wiring_embedding.go` | Gate ragService on rag.enabled |
| `internal/app/wiring_knowledge.go` | Gate ContextSearchAgent on rag.enabled |
| `internal/knowledge/store.go` | Two-step clamping for boost/decay |
