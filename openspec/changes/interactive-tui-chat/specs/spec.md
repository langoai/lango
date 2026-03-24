# Spec: Interactive TUI Chat

## Overview

Running `lango` without arguments starts an interactive terminal chat session. This replaces the WebUI Playground as the primary testing and interaction interface.

## Requirements

### Functional

1. `lango` (no args) starts interactive TUI chat
2. `lango serve` continues to work as before (full gateway + channels)
3. TUI streams agent responses in real-time via `TurnRunner.Run()`
4. Tool executions show inline approval prompts (a/s/d keys)
5. Slash commands: `/help`, `/clear`, `/new`, `/model`, `/status`, `/exit`, `/quit`
6. Chat history is scrollable (PgUp/PgDn)
7. Markdown responses are rendered via glamour on completion

### Non-Functional

1. Only Infra/Core/Buffer lifecycle components start (no network/automation overhead)
2. Graceful shutdown on Ctrl+D or double Ctrl+C
3. Context cancellation on Ctrl+C during streaming
4. Existing `lango serve` behavior is unchanged (regression-free)

## Key Bindings

| Key | State | Action |
|-----|-------|--------|
| Enter | idle | Send message |
| Alt+Enter | idle | Insert newline |
| Ctrl+C | streaming | Cancel generation |
| Ctrl+C | idle | Quit hint (double-tap to quit) |
| Ctrl+D | any | Immediate quit |
| a | approving | Allow tool execution |
| s | approving | Allow for session |
| d / Esc | approving | Deny tool execution |
| PgUp/PgDn | any | Scroll chat history |

## API Changes

### lifecycle.Registry

```go
func (r *Registry) SetMaxPriority(p Priority)
```

### app package

```go
type AppMode int
const AppModeServer AppMode = iota
const AppModeLocalChat

type AppOption func(*appOptions)
func WithLocalChat() AppOption

func New(boot *bootstrap.Result, opts ...AppOption) (*App, error)
```
