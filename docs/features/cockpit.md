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
- **Main content** -- active page rendering (chat, settings, tools, status, sessions, tasks, or approvals)
- **Context panel** -- live system metrics, toggled with `Ctrl+P`

## Pages

| Page | Description | Notes |
|------|-------------|-------|
| **Chat** | The primary AI conversation interface (same as `lango chat`) | Always has content |
| **Settings** | Interactive configuration viewer | Always has content |
| **Tools** | Tool inventory with agent assignments and invocation counts | Shows empty if ToolCatalog is nil |
| **Status** | System status dashboard (health, features, agent state) | Always has content |
| **Sessions** | Session history and management | Always has content |
| **Tasks** | Background task status and management | Shows empty if BackgroundManager is nil |
| **Approvals** | Approval history and grant management | Shows empty if ApprovalHistory is nil |

All 7 pages always appear in the sidebar navigation regardless of whether their backing data source is available. The chat page is the default active page on startup.

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
| `Ctrl+5` | Switch to Tasks page |
| `Ctrl+6` | Switch to Approvals page |

Sessions is accessible via sidebar navigation only (no keyboard shortcut).

## Context Panel

The context panel displays live system metrics in a right-side panel. It refreshes every 5 seconds via tick messages.

### Token Usage

Displays cumulative token counts for the session:

| Row | Description |
|-----|-------------|
| Input | Input tokens consumed |
| Output | Output tokens generated |
| Total | Combined total |
| Cache | Tokens served from cache |

### Tool Stats

Shows the top 5 tools by invocation count, sorted by frequency. Displays "No tool executions" when no tools have been called.

### Runtime Status (conditional)

Shown only while a turn is actively running (`IsRunning = true`). Hidden when idle.

- **Active agent indicator** -- green dot (●) with the currently active agent name
- **Delegation count** -- number of agent-to-agent delegations in the current turn (shown only when > 0)
- **Turn token count** -- token usage for the current turn (shown only when > 0)

### Channels (conditional)

Shown only when channel connections exist. Each channel displays:

- **Connection status** -- green dot (●) for connected, hollow circle (○) for disconnected
- **Channel name**
- **Message count** -- total messages received

### System

- **Uptime** -- process uptime since launch

When the observability collector is unavailable, placeholder text is shown.

## Cockpit vs Chat Mode

| Feature | Cockpit (`lango`) | Chat (`lango chat`) |
|---------|-------------------|---------------------|
| Sidebar navigation | Yes | No |
| Multiple pages | Yes (7 pages) | No (chat only) |
| Context panel | Yes | No |
| Keyboard shortcuts | Full set | Chat-only |
| Terminal width | Recommended 120+ cols | Any width |

## Tool Lifecycle Visibility

During streaming, each tool invocation appears as a distinct transcript item with lifecycle state:

- **Running** (⚙) — tool is executing
- **Success** (✓) — tool completed with duration
- **Error** (✗) — tool failed with error preview
- **Canceled** (⊘) — tool was canceled
- **Awaiting Approval** (🔒) — tool requires user approval

## Thinking Indicators

When the model uses extended thinking (via `genai.Part.Thought`), thinking phases appear as collapsible transcript items showing duration. A pending indicator (`⏳ Working...`) covers the submit-to-first-event gap.

## Two-Tier Approval

Approval requests are classified into two tiers based on tool safety level and capability:

- **Tier 1 (Inline Strip)** — compact single-line prompt for safe/moderate tools (e.g., browser_search, browser_observe)
- **Tier 2 (Fullscreen Dialog)** — overlay with risk badge, parameters, diff preview, and scroll for dangerous filesystem/exec tools (e.g., exec, fs_write, fs_edit)

Both tiers support the same actions: `a` (allow), `s` (allow session), `d`/`Esc` (deny).

### Double-Press Guardrail

Critical-risk tools (dangerous + filesystem or dangerous + automation) require pressing `a` or `s` twice within 3 seconds to confirm:

1. First press shows a warning: **"Press 'a' again to confirm (destructive operation)"**
2. Second press of the same key within 3 seconds executes the action
3. Pressing a different key or waiting longer than 3 seconds resets the pending state

In the inline strip, critical-risk tools are labeled with **(destructive)** in red next to the tool name.

### Rule Explanation

The fullscreen approval dialog includes a "Why: ..." explanation between the summary and parameters sections. This explanation is derived from the tool's SafetyLevel and Category combination:

| SafetyLevel + Category | Explanation |
|------------------------|-------------|
| dangerous + filesystem | "This tool modifies the filesystem and is classified as dangerous." |
| dangerous + automation | "This tool executes arbitrary code and is classified as dangerous." |
| moderate (any) | "This tool creates or modifies resources (moderate risk)." |
| other | "This tool requires approval under the current approval policy." |

## Runtime Visibility

During active turns, runtime events appear as inline items in the chat transcript:

### Delegation Events

Agent-to-agent delegation events display in the transcript as:

```
 🔀 from → to  reason
```

Where `from` and `to` are the agent names (highlighted), and the reason is shown in italics when provided.

### Budget Warnings

When delegation budget usage is reported, a warning appears:

```
⚠ Delegation budget: used/max (percentage%)
```

### Recovery Events

Recovery decisions during structured orchestration appear as:

```
 🔄 Action #attempt  (causeClass) backoff
```

The action label maps as follows:
- `retry` → "Retry"
- `retry_with_hint` → "Reroute"
- `direct_answer` → "Direct Answer"
- `escalate` → "Escalate"

### Turn Token Summary

After each assistant response completes, a token usage summary is appended to the transcript:

```
📊 Token usage: Xinput, Youtput, Ztotal (Wcached)
```

The cached portion is omitted when cache tokens are zero. Large numbers are formatted with k/M suffixes (e.g., "1.5k").

## Approvals Page

The Approvals page (`Ctrl+6`) provides a dedicated view for approval history and active session grants. Data refreshes automatically every 2 seconds.

### History Section

Displays a table of past approval decisions with the following columns:

| Column | Description |
|--------|-------------|
| Time | Relative timestamp (e.g., "2m ago") |
| Tool | Tool name that was evaluated |
| Summary | Brief description of the request |
| Outcome | Decision result (e.g., "granted", "denied", "bypass", "timeout") |
| Provider | Which approval provider handled the request |

### Grants Section

Shows currently active session-level grants:

| Column | Description |
|--------|-------------|
| Session | Session key identifier |
| Tool | Granted tool name |
| Granted | Relative time when the grant was created |

### Approvals Page Keys

| Key | Action |
|-----|--------|
| `/` | Toggle between history and grants sections |
| `↑`/`k` | Move cursor up |
| `↓`/`j` | Move cursor down |
| `r` | Revoke the selected grant (grants section only) |
| `R` | Revoke all grants for the selected session (grants section only) |

When both history and grants are empty, the page displays "No approval history yet."

## Background Tasks Page

The Tasks page (`Ctrl+5`) shows background tasks in a table view with columns for ID, Prompt, Status, and Elapsed time (elapsed is hidden on narrow terminals below 50 columns).

### Task Detail View

- **Enter** -- toggle the detail panel for the selected task
- **Esc** -- close the detail panel
- **↑/↓** -- scroll within the detail content when the detail panel is open

The detail panel shows:
- Status with elapsed time
- Origin channel (e.g., "telegram", "slack")
- Token usage
- Full prompt text (word-wrapped)
- Result text
- Error message (if any)

### Task Actions

When a `TaskActioner` is available:

| Key | Action | Applies to |
|-----|--------|------------|
| `c` | Cancel task | Running or pending tasks |
| `r` | Retry task | Failed or cancelled tasks |

Action results appear as a transient status message that auto-clears after 3 seconds.

## Background Task Strip

When a BackgroundManager is available, a compact task strip appears above the footer showing active task count and the most recent task's status. The full Tasks page (Ctrl+5) provides a detailed table view.

## Approval Operations

Approval handling is centered on the Chat page. When a tool call needs approval, the cockpit switches to Chat and renders either the inline strip or the fullscreen dialog depending on the tool's risk level. Operators respond with `a` to allow, `s` to allow for the session, or `d`/`Esc` to deny.

Critical-risk filesystem or automation tools still require the existing double-press confirmation before `a` or `s` takes effect. For the full policy model and approval-provider behavior, see [Tool Approval](../security/tool-approval.md) and [Approval CLI](../security/approval-cli.md).

## Channel Operations

With `--with-channels`, the cockpit acts as a live operator console for Telegram, Discord, and Slack. Channel messages flow into the Chat transcript through the EventBus, and approval requests from those sessions surface in the cockpit even when the operator is on another page.

Channel approvals apply to the originating channel session, so session grants and denials stay scoped to that remote conversation. Do not run `lango cockpit --with-channels` and `lango serve` against the same channel credentials at the same time. For setup details, see [Channels](channels.md).

## Background Task Operations

The Tasks page and the chat footer strip expose background task progress. Operators can inspect a task's detail panel, cancel running or pending work, and retry failed or cancelled tasks from within the cockpit. The page refreshes automatically, so the current task state stays visible without manual reloads.

For the system-level task model, CLI commands, and configuration reference, see [Background Tasks](../automation/background.md).

## Troubleshooting

Start with `lango doctor` when cockpit behavior looks wrong. The most common issues are:

- startup problems caused by a non-interactive terminal or unsupported alt-screen behavior
- empty context or runtime panels when the metrics collector has no data yet or no active turn is running
- missing channel messages or approval prompts when `--with-channels` is not enabled or channel credentials are invalid
- task actions that appear to do nothing because the selected task state does not allow retry or cancel

Check `~/.lango/cockpit.log` for the underlying error details when the TUI does not show enough context. Use a modern terminal with TTY and alt-screen support, and verify the channel/task wiring before assuming the cockpit UI is broken.
