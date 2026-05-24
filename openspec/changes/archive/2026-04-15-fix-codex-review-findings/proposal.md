## Why

Codex automated code review (3 rounds against `dev`) identified 14 bugs (4×P1, 10×P2) in the Phase 1-4 UX roadmap implementation on `feature/elastic-ux`. These range from trust-model gaps in extension pack integrity, to budget miscalculations in context injection, to plain-chat UX feature gaps. All have been fixed in-tree; this change documents the fixes and their spec-level impact.

## What Changes

### Round 1 (6 fixes — all applied)
- RAG budget split: recall and semantic RAG sections now share `budgets.RAG` (1/3 recall, 2/3 RAG) instead of each consuming the full budget
- Extension directory hash coverage: `fetchFromDir` hashes ALL files in skill directories (not just manifest-listed files)
- Extension prompt source wiring: `PromptSources()` now injected into prompt builder via `extensionPromptSections()`
- Retry partial stream guard: retry skipped when chunks already streamed to user (prevents garbled output)
- Plain chat session End: `runChat` now defers `Store.End(sessionKey)` for recall indexing
- view_skill ext path resolution: `SourcePack`-aware path resolution for extension-owned skills

### Round 2 (3 fixes — all applied)
- Stale-stream retry conflict: `staleTriggered` flag allows retry after stale timeout even if chunks were emitted
- --mode flag propagation: `chatCmd` reads inherited root persistent `--mode` flag and passes to `runChat(modeName)`
- Plain chat token usage: EventBus `TokenUsageEvent` subscription in `continuity_events.go` with `turnActive` pattern for `/cost`

### Round 3 (5 fixes — all applied)
- ext-* skill integrity enforcement: `AllowedExtPacks` filter on `FileSkillStore`, populated from extension registry's `OKPacks()`
- /clear session key rebinding: `SessionKey()` getter + live read in event closures + defer reordering
- Pack mirror full directory copy: `copyPackFiles` uses `copyTree` for skill directories (matches hash coverage)
- Directory-type skill path handling: `os.Stat` pre-check before `hashFile`, supports both file and directory manifest paths
- Emergency compaction with history: `totalMeasured` now includes `req.Contents` tokens and base prompt tokens

## Capabilities

### New Capabilities

(none — all fixes are to existing capabilities)

### Modified Capabilities
- `extension-pack-core`: hash coverage expanded to full skill directories; pack mirror now copies full directories; directory-type skill paths supported; `AllowedExtPacks` filter enforces integrity at skill-load time
- `inline-emergency-compaction`: trigger measurement now includes conversation history and base prompt tokens
- `session-recall`: plain chat path now calls `Store.End()` for recall indexing; `/clear` rebinds session key for continuity events
- `session-modes`: `--mode` flag propagated to plain chat via `chatCmd` → `runChat(modeName)`

## Impact

- `internal/adk/context_model.go` — RAG budget split + compaction measurement
- `internal/extension/source.go` — directory-aware hashing
- `internal/extension/installer.go` — full directory copy in pack mirror + skills store
- `internal/turnrunner/runner.go` — stale flag + partial stream guard
- `internal/skill/file_store.go` — `AllowedExtPacks` filter
- `internal/cli/chat/chat.go` — token accumulation, session key getter, turn reset
- `internal/cli/chat/continuity_events.go` — live session key + token event subscription
- `internal/app/wiring.go` — extension prompt injection + `extReg` in agentDeps
- `internal/app/wiring_extensions.go` — `extensionPromptSections()` function
- `internal/app/wiring_knowledge.go` — pass `extReg` to `initSkills`
- `internal/app/app.go` — `wireExtensionRegistry` moved before module build
- `internal/app/modules.go` — `extReg` field on `intelligenceModule`
- `internal/app/tools_meta.go` — `SourcePack`-aware `view_skill` path
- `cmd/lango/main.go` — mode propagation + session End defer reorder
