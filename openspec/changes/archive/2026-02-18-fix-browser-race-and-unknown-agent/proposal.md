## Why

Two runtime bugs discovered in Docker sidecar deployment: (1) browser tool panics with nil pointer dereference due to a race condition in `ensureBrowser()` — concurrent calls set `t.browser` without synchronization, and failed connections leave the field non-nil but disconnected; (2) ADK event replay logs "Event from an unknown agent: model" because the role-to-author fallback in `EventsAdapter.All()` maps unrecognized roles (including `"model"`) to the literal string `"model"`, which is not a registered agent name.

## What Changes

- Fix race condition in `browser.Tool.ensureBrowser()` using `sync.Once` with retry-on-failure semantics — only assign `t.browser` after successful `Connect()`
- Fix unknown agent "model" by adding `"model"` to the `"assistant"` case in the EventsAdapter role-to-author switch, and changing the default fallback to use `rootAgentName` instead of literal `"model"`

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `tool-browser`: Browser initialization is now thread-safe via `sync.Once`; failed connections are retried on next call instead of leaving browser in broken state
- `adk-architecture`: EventsAdapter role-to-author mapping now handles `"model"` role as assistant and defaults unknown roles to rootAgentName

## Impact

- **Files**: `internal/tools/browser/browser.go`, `internal/adk/state.go`
- **Concurrency**: Eliminates data race on `t.browser` field under concurrent tool invocations
- **Multi-agent**: Eliminates spurious "unknown agent" log warnings for model-generated events
