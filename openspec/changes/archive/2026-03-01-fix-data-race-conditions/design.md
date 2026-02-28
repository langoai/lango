## Context

Three packages fail under Go's `-race` detector due to unsynchronized concurrent access to shared memory:
1. **Slack mock**: handler goroutine appends to `PostMessages`/`UpdateMessages` slices while test goroutine reads them
2. **Telegram mock**: handler goroutine appends to `SentMessages`/`RequestCalls` slices while test goroutine reads them
3. **Exec background**: os/exec goroutine writes to `bytes.Buffer` while `GetBackgroundStatus()` reads via `String()`

## Goals / Non-Goals

**Goals:**
- Eliminate all data races detected by `-race` in the 3 affected packages
- Maintain backward-compatible API for `BackgroundProcess.Output` (same `String()` method)
- Keep fixes minimal and localized

**Non-Goals:**
- Refactoring test structure beyond what's needed for race safety
- Adding `-race` to CI pipeline (separate concern)
- Fixing potential races in other packages

## Decisions

### D1: Mutex per mock type (test code)
Add `sync.Mutex` directly to mock structs rather than using channels or atomic operations.
**Rationale**: Mutexes are the idiomatic Go pattern for protecting shared slices. Channels would overcomplicate simple append/read synchronization. Helper methods (`getPostMessages()`, etc.) return copies to prevent holding locks during assertions.

### D2: syncBuffer wrapper (production code)
Introduce a `syncBuffer` type that wraps `bytes.Buffer` with `sync.Mutex`, implementing `io.Writer` and `String()`.
**Rationale**: The buffer is used as `cmd.Stdout`/`cmd.Stderr` (requiring `io.Writer`) and read via `String()`. A thin wrapper keeps the change minimal. Alternatives considered:
- `sync.RWMutex`: unnecessary since writes are frequent and reads are infrequent
- External `io.Writer` with lock: would require changing `BackgroundProcess` API
- `bytes.Buffer` with external lock: callers would need to know about the lock

## Risks / Trade-offs

- [Minimal lock contention] → Background process output is written frequently but read rarely; mutex overhead is negligible
- [Type change on exported field] → `BackgroundProcess.Output` changes from `*bytes.Buffer` to `*syncBuffer`; no external consumers exist (internal package)
