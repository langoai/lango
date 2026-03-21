## 1. Shutdown Orchestration

- [x] 1.1 Refactor `lango serve` signal handling to support graceful shutdown on first signal and forced exit on second signal
- [x] 1.2 Add tests for the serve signal coordination helper and forced-exit behavior

## 2. Lifecycle Deadline Enforcement

- [x] 2.1 Update lifecycle shutdown coordination so per-component stop respects shutdown context and logs stop progress
- [x] 2.2 Update `app.Stop()` to surface lifecycle stop errors and preserve bounded shutdown behavior
- [x] 2.3 Add lifecycle tests covering blocked stop handlers, timeout behavior, and continued shutdown of later components

## 3. Context-Aware Stop Paths

- [x] 3.1 Change the internal channel shutdown contract to `Stop(ctx)` and update lifecycle registration accordingly
- [x] 3.2 Fix Telegram shutdown ordering and make Telegram/Slack/Discord stop paths return when shutdown context is done
- [x] 3.3 Change background manager and workflow engine shutdown APIs to honor context deadlines
- [x] 3.4 Add or update tests for channel stop behavior and manager shutdown timeout behavior

## 4. Documentation And Workflow Closure

- [x] 4.1 Update `openspec/specs/server/spec.md` and user-facing server documentation/README notes for the new shutdown behavior
- [x] 4.2 Run `go build ./...`, `go test ./...`, and OpenSpec verify/sync/archive steps for the completed change
