# Tasks: fix-codex-review-findings

All tasks have been implemented. This is a retroactive documentation of the 14 fixes applied across 3 Codex review rounds.

## Round 1: Initial findings (6 fixes)

- [x] 1.1 Split RAG budget between recall (1/3) and semantic (2/3) sections — `context_model.go`
- [x] 1.2 Hash all files in skill directory, not just manifest-listed file — `source.go`
- [x] 1.3 Wire `PromptSources()` into prompt builder via `extensionPromptSections()` — `wiring_extensions.go`, `wiring.go`
- [x] 1.4 Skip retry when chunks already streamed to user — `runner.go`
- [x] 1.5 Add `defer Store.End(sessionKey)` to `runChat` — `main.go`
- [x] 1.6 Resolve `view_skill` path using `SourcePack` for ext-owned skills — `tools_meta.go`

## Round 2: Conflict fixes and plain-chat gaps (3 fixes)

- [x] 2.1 Return `staleTriggered` flag from `wrapChunkCallbackWithStale`; allow retry when stale — `runner.go`
- [x] 2.2 Propagate `--mode` flag to `runChat(modeName)` with session pre-creation — `main.go`
- [x] 2.3 Subscribe to `TokenUsageEvent` in `continuity_events.go` with `turnActive` pattern — `chat.go`, `continuity_events.go`

## Round 3: Integrity enforcement and compaction (5 fixes)

- [x] 3.1 Add `AllowedExtPacks` filter to `FileSkillStore`; populate from `OKPacks()` — `file_store.go`, `wiring_knowledge.go`, `modules.go`
- [x] 3.2 Live session key read in event closures + `SessionKey()` getter + defer reorder — `chat.go`, `continuity_events.go`, `main.go`
- [x] 3.3 Copy full skill directories in pack mirror via `copyTree` — `installer.go`
- [x] 3.4 Handle directory-type skill paths with `os.Stat` pre-check — `source.go`
- [x] 3.5 Include `req.Contents` + base prompt tokens in emergency compaction measurement — `context_model.go`

## Deferred (separate change)

- [ ] D.1 Non-context builds need `ContextAwareModelAdapter` else-branch — `wiring.go`
- [ ] D.2 Local extension source snapshot before install — `source.go`
