## Why

Agent responses on Telegram appear to hang indefinitely. Logs reveal two contributing factors: (1) hallucinated sub-agent name retries with no success/failure logging, making it impossible to distinguish slow responses from actual hangs, and (2) a permanent per-session learning save rate limit (`maxLearnings=10`) that silently disables the learning system in long-running sessions. The rate limit provides no user-facing management mechanism, so once hit, learning is permanently lost for that session.

## What Changes

- **BREAKING**: Remove per-session rate limits (`maxKnowledge`, `maxLearnings`) from the knowledge store — these fields are removed from config and Store constructor
- Add learning data management tools (`learning_stats`, `learning_cleanup`) so the agent can brief users on stored learnings and clean up by criteria (age, confidence, category)
- Add new Store methods (`GetLearningStats`, `ListLearnings`, `DeleteLearning`, `DeleteLearningsWhere`) for learning data lifecycle management
- Add comprehensive timing and result logging to ADK agent `RunAndCollect` (hallucination retry success/failure, elapsed time)
- Add request-level observability to `runAgent` in channels (start/complete/fail/timeout logs, 80% timeout early warning)
- Remove `Max Learnings` and `Max Knowledge` fields from CLI/TUI settings forms

## Capabilities

### New Capabilities
- `learning-data-management`: Agent tools and store methods for learning data statistics, listing, and cleanup (single/bulk delete by criteria)

### Modified Capabilities
- `knowledge-store`: Remove per-session rate limiting (`maxKnowledge`, `maxLearnings` fields and `reserveXxxSlot` functions), change `NewStore` constructor signature
- `agent-runtime`: Add timing/result logging to `RunAndCollect` for hallucination retry observability
- `learning-engine`: Enhance error logging with session/tool context
- `cli-settings`: Remove `Max Learnings` and `Max Knowledge` form fields from Knowledge settings

## Impact

- **Config**: `KnowledgeConfig.MaxLearnings` and `KnowledgeConfig.MaxKnowledge` fields removed — existing config files with these fields will be silently ignored by viper
- **Store API**: `knowledge.NewStore()` signature changes from `(client, logger, maxKnowledge, maxLearnings)` to `(client, logger)` — all callers must update
- **Agent Tools**: Two new tools registered in `buildMetaTools` — `learning_stats` (safe) and `learning_cleanup` (moderate safety)
- **Logging**: New structured log entries in `internal/adk/agent.go` and `internal/app/channels.go` with timing data
- **TUI**: Knowledge settings form loses two fields; `state_update.go` loses two config update cases
