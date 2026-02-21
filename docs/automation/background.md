# Background Tasks

!!! warning "Experimental"
    The background task system is experimental. APIs and behavior may change in future releases.

In-memory background task manager for asynchronous agent operations. Submit long-running prompts to execute in the background while continuing to interact with the agent.

## Features

### Concurrency Limiting

A semaphore controls how many tasks run simultaneously. When the limit is reached, new submissions are rejected with an error rather than queued.

### Task State Machine

Each task follows a strict lifecycle with mutex-protected transitions:

```
Pending --> Running --> Done
                   --> Failed
                   --> Cancelled
```

| State | Description |
|-------|-------------|
| `pending` | Task created, waiting for a semaphore slot |
| `running` | Agent is actively processing the prompt |
| `done` | Execution completed successfully |
| `failed` | Execution encountered an error |
| `cancelled` | Task was cancelled by the user |

### Completion Notifications

When a task finishes (success or failure), results are automatically delivered to the channel that initiated the request. A typing indicator is shown while the agent processes the prompt.

### Monitoring

The manager tracks all tasks in memory, providing list and status queries for active task counts and summaries.

## CLI Commands

### Submit a Task

```bash
lango bg run --prompt "Analyze the codebase for security vulnerabilities"
```

### List Tasks

```bash
lango bg list
```

### Check Status

```bash
lango bg status --id <task-id>
```

### Get Result

```bash
lango bg result --id <task-id>
```

### Cancel a Task

```bash
lango bg cancel --id <task-id>
```

## Ephemeral Storage

!!! note "In-Memory Only"
    Background tasks are stored in memory only. All task state is lost when the application restarts. For persistent scheduled execution, use the [Cron](cron.md) system instead.

Each task runs in an isolated session with the key format `bg:<task-id>`.

## Configuration

```yaml
background:
  enabled: true
  yieldMs: 5000               # Auto-yield threshold in milliseconds
  maxConcurrentTasks: 10       # Max parallel tasks (default: 10)
  taskTimeout: 30m             # Max duration per task (default: 30m)
  defaultDeliverTo:            # Fallback delivery channels
    - telegram
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `background.enabled` | `bool` | `false` | Enable the background task system |
| `background.yieldMs` | `int` | - | Time in ms before auto-yielding to background |
| `background.maxConcurrentTasks` | `int` | `10` | Maximum concurrently running tasks |
| `background.taskTimeout` | `duration` | `30m` | Maximum duration for a single task |
| `background.defaultDeliverTo` | `[]string` | `[]` | Default delivery channels |

## Architecture

The background system consists of two main components:

- **Manager** (`internal/background/manager.go`) -- handles task lifecycle, concurrency limiting, and execution
- **Task** (`internal/background/task.go`) -- represents a single execution unit with thread-safe state transitions
