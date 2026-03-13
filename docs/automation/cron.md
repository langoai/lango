# Cron Scheduling

Persistent cron scheduling system powered by [robfig/cron/v3](https://github.com/robfig/cron) with Ent ORM storage. Jobs survive application restarts -- on startup, all enabled jobs are loaded from the database and re-registered with the scheduler.

## Schedule Types

| Type | Flag | Example | Description |
|------|------|---------|-------------|
| `cron` | `--schedule` | `"0 9 * * *"` | Standard cron expression (minute, hour, day, month, weekday) |
| `every` | `--every` | `1h` | Interval-based repetition using Go duration syntax |
| `at` | `--at` | `2026-02-20T15:00:00Z` | One-time execution at a specific RFC3339 datetime |

One-time (`at`) jobs are automatically disabled after execution.

## CLI Commands

### Add a Cron Job

```bash
# Daily summary at 9 AM
lango cron add --name "daily-summary" \
  --schedule "0 9 * * *" \
  --prompt "Summarize yesterday's activity" \
  --deliver-to telegram

# Every 2 hours
lango cron add --name "health-check" \
  --every 2h \
  --prompt "Check all systems status" \
  --deliver-to slack

# One-time execution
lango cron add --name "deploy-reminder" \
  --at "2026-02-20T15:00:00Z" \
  --prompt "Remind team about the deployment window"
```

### Per-Job Timeout

Each job can specify a per-job timeout that overrides the global `cron.defaultJobTimeout`. If no per-job timeout is set, the global default (30 minutes) applies.

```bash
# 5-minute timeout for a quick check
lango cron add --name "quick-check" \
  --every 30m \
  --prompt "Check API latency" \
  --timeout 5m

# 2-hour timeout for a long-running report
lango cron add --name "monthly-report" \
  --schedule "0 2 1 * *" \
  --prompt "Generate full monthly report" \
  --timeout 2h
```

The `timeout` parameter accepts Go duration syntax (e.g., `5m`, `1h30m`, `2h`). Precedence:

1. **Per-job `--timeout`** — if set on the job, this value is used
2. **Global `cron.defaultJobTimeout`** — fallback when the job has no explicit timeout (default: `30m`)

!!! note "Idempotent Upsert"
    Re-adding a job with the same `--name` updates the existing job instead of creating a duplicate. This makes it safe to re-run `cron add` commands — the schedule, prompt, timeout, and delivery settings are updated in place.

### List Jobs

```bash
lango cron list
```

### Pause / Resume

```bash
lango cron pause --id <job-id>
lango cron resume --id <job-id>
```

### Delete a Job

```bash
lango cron delete --id <job-id>
```

### View History

```bash
# History for a specific job
lango cron history --id <job-id> --limit 10

# History across all jobs
lango cron history --limit 20
```

## Session Modes

Each cron job runs in its own agent session. The session mode controls whether conversations persist across runs:

| Mode | Session Key Format | Behavior |
|------|-------------------|----------|
| `isolated` (default) | `cron:<name>:<timestamp>` | Fresh session every execution. No memory of previous runs. |
| `main` | `cron:<name>` | Shared session across all runs. Agent remembers previous outputs. |

```bash
# Use shared session (agent remembers previous runs)
lango cron add --name "weekly-report" \
  --schedule "0 9 * * 1" \
  --prompt "Write this week's report, building on previous ones" \
  --isolated=false
```

## Result Delivery

Job results are delivered to configured communication channels after execution. If no `deliver_to` is specified per-job, the system falls back to `cron.defaultDeliverTo` from the configuration.

!!! warning "No Delivery Channel"
    If no delivery channel is configured (neither per-job nor default), job results are logged but not delivered to any channel. A warning is emitted in the logs.

## Configuration

> **Settings:** `lango settings` → Cron Scheduler

```json
{
  "cron": {
    "enabled": true,
    "timezone": "Asia/Seoul",
    "maxConcurrentJobs": 5,
    "defaultSessionMode": "isolated",
    "defaultJobTimeout": "30m",
    "historyRetention": "30d",
    "defaultDeliverTo": ["telegram"]
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `cron.enabled` | `bool` | `false` | Enable the cron scheduling system |
| `cron.timezone` | `string` | `"UTC"` | Default timezone for cron expressions |
| `cron.maxConcurrentJobs` | `int` | `5` | Maximum concurrently executing jobs |
| `cron.defaultSessionMode` | `string` | `"isolated"` | Default session mode for new jobs |
| `cron.historyRetention` | `string` | - | Duration to retain execution history |
| `cron.defaultJobTimeout` | `string` | `"30m"` | Default timeout for job execution (Go duration) |
| `cron.defaultDeliverTo` | `[]string` | `[]` | Default delivery channels |

## Architecture

The cron system consists of three main components:

- **Scheduler** (`internal/cron/scheduler.go`) -- manages job registration, lifecycle, and the concurrency semaphore
- **Executor** (`internal/cron/executor.go`) -- runs individual jobs via `AgentRunner`, persists history, and delivers results
- **Store** (`internal/cron/store.go`) -- Ent ORM persistence layer for jobs and execution history
