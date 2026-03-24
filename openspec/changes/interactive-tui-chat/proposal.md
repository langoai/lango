# Interactive TUI Chat

## Problem

Lango requires `lango serve` to interact with the agent, which starts a full gateway server with HTTP/WebSocket endpoints and all automation/network components. There is no lightweight, terminal-native way to chat with the agent directly.

## Solution

Add an interactive TUI chat mode as the default entry point (`lango` with no arguments). This provides a Claude Code-like terminal experience with:

- Streaming agent responses with markdown rendering (glamour)
- Inline tool approval (a/s/d keys)
- Slash commands (/help, /clear, /model, /status, /exit)
- Mode-aware component lifecycle (only core components start, not network/automation)

## Scope

- Remove WebUI Playground (replaced by native TUI)
- Add `Registry.SetMaxPriority()` for mode-aware lifecycle filtering
- Add `app.WithLocalChat()` option for lightweight app initialization
- New `internal/cli/chat/` package (8 files) implementing bubbletea chat model
- Wire `runChat()` into `cmd/lango/main.go` as root command RunE
- Update all documentation (README, CLI docs, gateway docs)

## Non-goals

- Hot-reload model switching (`/model [name]` change)
- Session compaction (`/compact`)
- Settings editing within TUI (`/settings`)
