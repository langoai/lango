# Background Tasks

!!! warning "Experimental"
    The background task system is experimental. APIs and behavior may change in future releases.

In-memory background task manager for asynchronous agent operations. Submit long-running prompts to execute in the background while continuing to interact with the agent.

## Features

### Concurrency Limiting

A semaphore controls how many tasks run simultaneously. When the limit is reached, new submissions are rejected with an error rather than queued. The semaphore size defaults to `maxConcurrentTasks`.

### Task State Machine

Each task follows a strict lifecycle with mutex-protected transitions:

```
Pending --> Running --> Done
                   \-> Failed
                   \-> Cancelled
```

| State | Description |
|-------|-------------|
| `pending` | Task created, waiting for a semaphore slot |
| `running` | Agent is actively processing the prompt |
| `done` | Execution completed successfully |
| `failed` | Execution encountered an error |
| `cancelled` | Task was cancelled by the user |

#### State Transition Methods

| Method | Transition | Side effects |
|--------|-----------|-------------|
| `SetRunning()` | pending -> running | Records `StartedAt` timestamp |
| `Complete(result)` | running -> done | Records result and `CompletedAt` timestamp |
| `Fail(errMsg)` | running -> failed | Records error message and `CompletedAt` timestamp |
| `Cancel()` | pending/running -> cancelled | Records `CompletedAt` timestamp and invokes cancel function |

All transitions are protected by a `sync.RWMutex`. The `Cancel` method also invokes the task's context cancel function to stop the running agent.

#### TaskSnapshot

A `TaskSnapshot` is an immutable copy of a task's state, safe for concurrent reading. It includes:

- `ID`, `Status`, `StatusText` -- task identity and current state
- `Prompt`, `Result`, `Error` -- input prompt, output result, or error message
- `OriginChannel`, `OriginSession` -- where the task was initiated
- `StartedAt`, `CompletedAt` -- timing information
- `TokensUsed` -- token count consumed during execution

### Completion Notifications

When a task finishes (success or failure), results are automatically delivered to the channel that initiated the request. The notification system provides:

- **Start notifications** -- sent when a task begins execution with a truncated prompt summary
- **Typing indicators** -- shown on the origin channel while the agent processes the prompt
- **Completion notifications** -- formatted differently based on terminal state (done, failed, cancelled)

If no origin channel is set, notifications are skipped with a warning suggesting `background.defaultDeliverTo` in settings.

### Tool Approval Routing

When a background task runs, tool approval requests are routed to the originating session or channel. This ensures approval prompts reach the user who submitted the task, not a different session.

### Monitoring

The `Monitor` component tracks all tasks in memory, providing:

- `ActiveCount()` -- number of tasks currently pending or running
- `Summary()` -- aggregate counts by state (total, pending, running, done, failed, cancelled)

## Agent Tools

Background tasks are submitted through agent tools, not CLI commands. The agent can invoke these tools during a conversation:

| Tool | Description |
|------|-------------|
| `bg_submit` | Submit a prompt for asynchronous background execution |
| `bg_status` | Check the status of a background task |
| `bg_list` | List all background tasks and their current status |
| `bg_result` | Retrieve the result of a completed background task |
| `bg_cancel` | Cancel a pending or running background task |

### bg_submit Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `prompt` | Yes | The prompt to execute in the background |
| `channel` | No | Channel to deliver results to (e.g. `telegram:CHAT_ID`) |

Channel auto-detection: if `channel` is omitted, the tool attempts to detect the delivery target from the session context. If that also fails, the `background.defaultDeliverTo` config value is used as a fallback.

Each background task runs in an isolated session with the key format `bg:<task-id>`.

## CLI Commands

The CLI provides read-only management commands for background tasks. Task submission is handled exclusively through agent tools.

### List Tasks

```bash
lango bg list
```

Displays a table with columns: ID (truncated to 8 chars), STATUS, PROMPT (truncated to 50 chars), STARTED, DURATION.

### Check Status

```bash
lango bg status <id>
```

Shows full task details including origin channel, session, timing, error messages, and result.

### Get Result

```bash
lango bg result <id>
```

Returns the result text of a completed task. Fails if the task is not in `done` state.

### Cancel a Task

```bash
lango bg cancel <id>
```

Cancels a task that is `pending` or `running`. Fails if the task is already in a terminal state.

## Ephemeral Storage

!!! note "In-Memory Only"
    Background tasks are stored in memory only. All task state is lost when the application restarts. For persistent scheduled execution, use the [Cron](cron.md) system instead.

## Shutdown

When the application shuts down, `Manager.Shutdown()` cancels all pending and running tasks to ensure clean resource release.

## Configuration

> **Settings:** `lango settings` -> Background Tasks

```json
{
  "background": {
    "enabled": true,
    "yieldMs": 5000,
    "maxConcurrentTasks": 10,
    "taskTimeout": "30m",
    "defaultDeliverTo": ["telegram"]
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `background.enabled` | `bool` | `false` | Enable the background task system |
| `background.yieldMs` | `int` | - | Time in ms before auto-yielding to background |
| `background.maxConcurrentTasks` | `int` | `10` | Maximum concurrently running tasks |
| `background.taskTimeout` | `duration` | `30m` | Maximum duration for a single task |
| `background.defaultDeliverTo` | `[]string` | `[]` | Default delivery channels |

## Architecture

The background system consists of four components:

- **Manager** (`internal/background/manager.go`) -- handles task lifecycle, concurrency limiting, submission, and execution. Uses a channel-based semaphore for concurrency control and `context.WithTimeout` for task timeout enforcement.
- **Task** (`internal/background/task.go`) -- represents a single execution unit with thread-safe state transitions and immutable snapshot reads.
- **Notification** (`internal/background/notification.go`) -- handles sending start, completion, and failure notifications to the origin channel. Manages typing indicators during execution.
- **Monitor** (`internal/background/monitor.go`) -- provides aggregate task state summaries and active task counts for observability.
