# Tasks: Interactive TUI Chat

## Implementation

- [x] Add `Registry.SetMaxPriority()` to lifecycle registry with tests
- [x] Add `AppMode`, `AppOption`, `WithLocalChat()` to app package
- [x] Update `app.New()` to accept options and filter lifecycle by mode
- [x] Skip channels init and post-build lifecycle in LocalChat mode
- [x] Remove Playground WebUI (playground.go, playground/index.html, route)
- [x] Add `glamour` dependency
- [x] Create `internal/cli/chat/messages.go` (tea.Msg types)
- [x] Create `internal/cli/chat/markdown.go` (glamour rendering)
- [x] Create `internal/cli/chat/approval.go` (TUI approval provider)
- [x] Create `internal/cli/chat/input.go` (textarea wrapper)
- [x] Create `internal/cli/chat/chatview.go` (scrollable viewport)
- [x] Create `internal/cli/chat/statusbar.go` (status + help bars)
- [x] Create `internal/cli/chat/commands.go` (slash commands)
- [x] Create `internal/cli/chat/chat.go` (root bubbletea model)
- [x] Wire `runChat()` into `cmd/lango/main.go` root command RunE
- [x] Override TTY fallback with TUI approval provider

## Tests

- [x] `TestRegistry_SetMaxPriority_SkipsHighPriority`
- [x] `TestRegistry_SetMaxPriority_RollbackOnFailure`

## Documentation

- [x] Update README.md (default entry point, CLI commands table)
- [x] Update docs/cli/core.md (add `lango` TUI section)
- [x] Update docs/cli/index.md (quick reference table)
- [x] Update docs/getting-started/quickstart.md (replace playground tip)
- [x] Remove playground section from docs/gateway/http-api.md

## Verification

- [x] `go build ./...` passes
- [x] `go test ./...` passes
- [x] `go mod tidy` clean
