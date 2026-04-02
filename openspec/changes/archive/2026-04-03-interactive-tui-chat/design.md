# Design: Interactive TUI Chat

## Architecture

### Mode-Aware Lifecycle

```
Registry.SetMaxPriority(PriorityBuffer)  // 300
├── PriorityInfra(100)     → STARTED (session store, security)
├── PriorityCore(200)      → STARTED (supervisors, core services)
├── PriorityBuffer(300)    → STARTED (embedding, memory buffers)
├── PriorityNetwork(400)   → SKIPPED (gateway, P2P, MCP, channels)
└── PriorityAutomation(500)→ SKIPPED (cron, background, workflow)
```

### Component Stack

```
cmd/lango/main.go          runChat() entry point
    ↓
internal/cli/chat/         bubbletea TUI (UI layer)
    ↓
internal/turnrunner/       shared execution runtime
    ↓
internal/adk/              agent runtime (ADK)
```

### Approval Flow

```
TurnRunner.Run()
    → WithApproval middleware
        → CompositeProvider.RequestApproval()
            → (no Gateway/Channel provider matches)
            → TTY fallback = TUIApprovalProvider
                → program.Send(ApprovalRequestMsg)
                → user presses a/s/d
                → response sent via channel
```

### File Structure

```
internal/cli/chat/
├── chat.go        Root bubbletea model (ChatModel)
├── messages.go    tea.Msg types (ChunkMsg, DoneMsg, etc.)
├── commands.go    Slash command registry
├── input.go       textarea wrapper
├── chatview.go    Scrollable chat viewport
├── statusbar.go   Top status bar + bottom help bar
├── markdown.go    glamour-based rendering
└── approval.go    TUI approval provider + banner
```

## Key Decisions

1. **TurnRunner reuse**: TUI calls `app.TurnRunner.Run()` directly — gets timeout, trace, approval state, sanitizer for free.
2. **Approval override**: Replace TTY fallback on existing CompositeProvider rather than creating a new one. Minimal invasiveness.
3. **No new lifecycle components**: TUI runs in the main goroutine via `tea.Program.Run()`. No lifecycle registration needed.
4. **Alt screen**: Uses `tea.WithAltScreen()` for clean terminal experience.
