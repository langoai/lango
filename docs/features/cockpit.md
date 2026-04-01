---
title: Cockpit TUI
---

# Cockpit TUI

## Overview

The cockpit is a multi-panel terminal dashboard and the default entry point when running `lango` with no arguments. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), it wraps the chat interface in a full-featured layout with sidebar navigation, multiple pages, and a live context panel.

## Launch

```bash
lango            # launch cockpit (default)
lango cockpit    # explicit cockpit launch
lango chat       # plain single-panel chat (no sidebar, no pages)
```

The cockpit requires an interactive terminal with TTY support.

## Layout

```
┌──────────┬─────────────────────────┬──────────────┐
│          │                         │              │
│ Sidebar  │     Main Content        │   Context    │
│ (pages)  │     (active page)       │   Panel      │
│          │                         │  (metrics)   │
│          │                         │              │
└──────────┴─────────────────────────┴──────────────┘
```

- **Sidebar** -- page navigation list, toggled with `Ctrl+B`
- **Main content** -- active page rendering (chat, settings, tools, status, or sessions)
- **Context panel** -- live system metrics, toggled with `Ctrl+P`

## Pages

| Page | Description |
|------|-------------|
| **Chat** | The primary AI conversation interface (same as `lango chat`) |
| **Settings** | Interactive configuration viewer |
| **Tools** | Tool inventory with agent assignments and invocation counts |
| **Status** | System status dashboard (health, features, agent state) |
| **Sessions** | Session history and management |

The chat page is always the default active page on startup.

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Ctrl+B` | Toggle sidebar visibility |
| `Ctrl+P` | Toggle context panel |
| `Tab` | Switch focus between sidebar and main content |
| `Ctrl+Y` | Copy to clipboard |
| `Ctrl+1` | Switch to Chat page |
| `Ctrl+2` | Switch to Settings page |
| `Ctrl+3` | Switch to Tools page |
| `Ctrl+4` | Switch to Status page |

## Context Panel

The context panel displays live system metrics from the observability collector:

- Goroutine count and memory usage
- Per-tool invocation counts (sorted by frequency)
- Session metrics
- Process uptime

The panel refreshes periodically via tick messages. When the observability collector is unavailable, placeholder text is shown.

## Cockpit vs Chat Mode

| Feature | Cockpit (`lango`) | Chat (`lango chat`) |
|---------|-------------------|---------------------|
| Sidebar navigation | Yes | No |
| Multiple pages | Yes (5 pages) | No (chat only) |
| Context panel | Yes | No |
| Keyboard shortcuts | Full set | Chat-only |
| Terminal width | Recommended 120+ cols | Any width |
