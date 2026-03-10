## Context

The P2P workspace & git integration (implemented in change `2026-03-10-p2p-workspace-git`) added 13 new files across 4 packages. A three-pronged code review (reuse, quality, efficiency) identified concurrency bugs, dead code, redundant abstractions, and inefficient patterns that need cleanup.

## Goals / Non-Goals

**Goals:**
- Fix the `Node.PubSub()` data race (concurrent callers could create duplicate GossipSub instances)
- Remove dead code paths (chronicler no-op wiring, dead `git fetch` subprocess)
- Eliminate redundant types and stringly-typed fields
- Improve protocol handler memory efficiency for large bundles
- Extract shared error sentinels to reduce string duplication

**Non-Goals:**
- Changing workspace/git feature behavior or adding new capabilities
- Refactoring CLI commands beyond error sentinel extraction
- Adding tests for existing untested packages

## Decisions

1. **`sync.Once` for PubSub lazy init** — Protects against concurrent callers creating duplicate GossipSub instances. Simpler and more idiomatic than mutex + double-check locking.

2. **Remove dead chronicler block in app.go** — The code converted workspace triples to a local struct then discarded them (`_ = gs`). Removing it leaves the nil-adder chronicler from wiring, which correctly short-circuits via `addTriples == nil`.

3. **Build WorkspaceGossip once** — Move chronicler/tracker creation before gossip, then construct gossip once with the handler already set. Eliminates the wasteful first construction.

4. **`json.NewDecoder` for streaming** — Replaces `io.ReadAll` + `json.Unmarshal` with streaming decode. For a 50MB bundle push, this avoids holding two copies in memory.

5. **Sentinel errors** — `errLimitReached` replaces fragile `err.Error() == "limit reached"` string comparison. `errP2PDisabled` eliminates 10+ identical error string literals in CLI.

6. **`Role` type** — Follow the existing `Status` string enum pattern in the same file. Prevents typos in role assignment.

7. **Merge `workspaceComponents` into `wsComponents`** — Both types are in package `app`. The former is a strict subset of the latter. `buildWorkspaceTools` simply ignores the extra fields.

## Risks / Trade-offs

- [Low] `sync.Once` caches errors permanently — if PubSub creation fails once, all future calls fail too → Acceptable; PubSub creation failure is fatal anyway and the node should not continue.
- [Low] Streaming JSON decoder loses the ability to retry on partial reads → Protocol already has a 5-minute timeout and LimitReader; retries are handled at a higher level.
