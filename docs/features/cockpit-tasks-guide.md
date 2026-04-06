# Cockpit Background Task Operator Guide

This guide covers operating background tasks from the cockpit TUI. For the
canonical reference on the background task system (architecture, agent tools,
notifications, configuration), see [docs/automation/background.md](../automation/background.md).

## Overview

Background tasks let you offload long-running agent prompts so the main
conversation remains responsive. The cockpit integrates background tasks in two
places:

1. **Tasks page** -- a dedicated table view with detail panel, accessible via
   the sidebar or `Ctrl+5`.
2. **Task strip** -- a compact footer bar embedded in the chat view, showing
   live counts and the most recent task.

Both update automatically every few seconds while active.

## Task Lifecycle

Every background task follows a five-state lifecycle:

```
Pending --> Running --> Done
                   \-> Failed
                   \-> Cancelled
```

| State | Meaning |
|-------|---------|
| `pending` | Created, waiting for a concurrency slot |
| `running` | Agent is actively executing the prompt |
| `done` | Completed successfully; result is available |
| `failed` | Execution encountered an error |
| `cancelled` | Cancelled by the user or by shutdown |

Transitions are mutex-protected and irreversible -- a cancelled task cannot
be overwritten by a late completion. For transition details, see
[background.md -- Task State Machine](../automation/background.md#task-state-machine).

## Submitting Tasks

Background tasks are submitted through **agent tools**, not the cockpit UI
directly. During a conversation the agent can invoke:

- `bg_submit` -- submit a prompt for asynchronous execution

The cockpit chat view works as the conversation surface where you interact with
the agent. When the agent calls `bg_submit`, the task appears in the Tasks page
and task strip immediately.

See [background.md -- Agent Tools](../automation/background.md#agent-tools) for
the full tool reference and parameters.

## Cockpit Tasks Page

Open the Tasks page by clicking "Tasks" in the sidebar or pressing `Ctrl+5`.

### Table View

The table displays all background tasks with these columns:

| Column | Description |
|--------|-------------|
| ID | Truncated task identifier |
| Prompt | Task prompt text (truncated to fit) |
| Status | Current lifecycle state |
| Elapsed | Time since start, or total duration if terminal |

When the terminal width is below 50 columns, the Elapsed column is hidden
to preserve readability.

### Keyboard Navigation

| Key | Action |
|-----|--------|
| `Up` / `k` | Move cursor up in table (or scroll up in detail) |
| `Down` / `j` | Move cursor down in table (or scroll down in detail) |
| `Enter` | Toggle detail panel for the selected task |
| `Esc` | Close the detail panel |
| `c` | Cancel the selected task (pending/running only) |
| `r` | Retry the selected task (failed/cancelled only) |

The task list refreshes every 2 seconds while the page is active.

### Task Count

A title bar reading "Background Tasks" appears at the top of the page. The
table header shows column labels, and all tasks are listed below a separator
line. When no tasks exist, the page displays "No active tasks".

## Detail Panel

Press `Enter` on a selected task to expand the detail panel below the table.

### Displayed Fields

| Field | Description |
|-------|-------------|
| Status | Current state with elapsed time |
| Origin | Channel that initiated the task (e.g., `telegram`, `slack`), or `(none)` |
| Tokens | Token count consumed during execution |
| Prompt | Full prompt text, word-wrapped |
| Result | Completion result text, or `(none)` if not yet done |
| Error | Error message if failed, or `(none)` |

### Height Clamp

When the total terminal height is 14 rows or more, the detail panel is
allocated up to 60% of the height and the table retains at least 40%. The
table always gets a minimum of 6 rows and the detail panel a minimum of 8
rows.

### Scrolling

Use `Up`/`k` and `Down`/`j` to scroll through the detail content when it
exceeds the available height.

## Task Actions

### Cancel

Press `c` to cancel the currently selected task. This is only available when
the task status is `pending` or `running`. Cancellation invokes the task's
context cancel function, which aborts the running agent.

### Retry

Press `r` to retry a `failed` or `cancelled` task. Retry re-submits the
original prompt with the same origin channel and session, creating a new task.

### Feedback

Both actions produce an asynchronous status message displayed in the task page
with a 3-second TTL. On success the message reads "Cancelled: <id>" or
"Retried: <id>". On error the message shows "Error: <details>".

## CLI Commands

The `lango bg` subcommands provide the same task management outside the TUI.
These require a running server (`lango serve`).

| Command | Description |
|---------|-------------|
| `lango bg list` | Table view of all tasks (ID, status, prompt, started, duration) |
| `lango bg status <id>` | Full details for a single task |
| `lango bg cancel <id>` | Cancel a pending or running task |
| `lango bg result <id>` | Print the result of a completed task |

For detailed output format, see
[background.md -- CLI Commands](../automation/background.md#cli-commands).

## Task Strip

The chat view (both `lango chat` and the cockpit Chat page) includes a
compact task strip footer bar. It renders when at least one background task
exists and shows:

- **Running count** -- number of currently running tasks
- **Pending count** -- shown only when > 0
- **Latest task** -- name (truncated to 30 chars), status, and elapsed time

The strip refreshes periodically alongside the chat tick cycle. When no
background manager is configured or no tasks exist, the strip is hidden.

## Configuration

Background task settings (max concurrent tasks, timeout, default delivery
channels) are configured through the application config. See
[background.md -- Configuration](../automation/background.md#configuration)
for the full reference.

Key settings:

| Setting | Default | Purpose |
|---------|---------|---------|
| `background.maxConcurrentTasks` | `10` | Max concurrent non-terminal tasks |
| `background.taskTimeout` | `30m` | Maximum duration per task |
| `background.defaultDeliverTo` | `[]` | Default notification channels |
