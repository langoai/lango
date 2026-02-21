# Automation Commands

Commands for managing cron jobs, and workflow pipelines. See the [Automation](../automation/index.md) section for detailed documentation.

---

## Cron Commands

Manage scheduled cron jobs that execute agent prompts on a recurring or one-time basis. Cron must be enabled in configuration (`cron.enabled = true`).

```
lango cron <subcommand>
```

### lango cron add

Add a new scheduled cron job. Exactly one scheduling method must be specified: `--schedule`, `--every`, or `--at`.

```
lango cron add --name <name> --prompt <text> [scheduling flags] [options]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--name` | string | *required* | Job name |
| `--prompt` | string | *required* | Prompt to execute |
| `--schedule` | string | | Cron expression (e.g., `"0 9 * * *"`) |
| `--every` | string | | Interval (e.g., `"1h"`, `"30m"`) |
| `--at` | string | | One-time execution (ISO8601: `"2026-02-20T15:00:00"`) |
| `--deliver` | strings | | Channels to deliver results (e.g., `slack,telegram`) |
| `--isolated` | bool | `false` | Run in isolated session |
| `--timezone` | string | `UTC` | Timezone for scheduling |

!!! note "Scheduling Methods"
    Exactly one of `--schedule`, `--every`, or `--at` is required. They cannot be combined.

    - `--schedule` uses standard cron syntax (5 fields: minute, hour, day-of-month, month, day-of-week)
    - `--every` uses Go duration format (e.g., `30m`, `1h`, `2h30m`)
    - `--at` uses ISO8601 format for one-time execution

**Examples:**

```bash
# Daily at 9 AM UTC
$ lango cron add \
    --name "daily-news" \
    --schedule "0 9 * * *" \
    --prompt "Summarize today's top tech news" \
    --deliver slack
Cron job "daily-news" created (id: a1b2c3d4)
  Schedule: cron 0 9 * * *
  Prompt: Summarize today's top tech news
  Deliver to: [slack]

# Every hour
$ lango cron add \
    --name "health-check" \
    --every 1h \
    --prompt "Check all server endpoints and report any issues" \
    --isolated

# One-time scheduled execution
$ lango cron add \
    --name "meeting-prep" \
    --at "2026-02-25T15:00:00" \
    --prompt "Prepare meeting notes for the Q1 review" \
    --timezone "America/New_York"
```

---

### lango cron list

List all registered cron jobs with their status and schedule.

```
lango cron list
```

**Output columns:**

| Column | Description |
|--------|-------------|
| ID | Short job ID (first 8 characters) |
| NAME | Job name |
| SCHEDULE | Schedule type and expression |
| ENABLED | `yes` or `no` |
| LAST RUN | Last execution timestamp (or `-`) |
| NEXT RUN | Next scheduled execution (or `-`) |

**Example:**

```bash
$ lango cron list
ID        NAME           SCHEDULE           ENABLED  LAST RUN              NEXT RUN
a1b2c3d4  daily-news     cron 0 9 * * *     yes      2026-02-20 09:00:00   2026-02-21 09:00:00
e5f6g7h8  health-check   every 1h           yes      2026-02-20 14:00:00   2026-02-20 15:00:00
i9j0k1l2  meeting-prep   at 2026-02-25...   yes      -                     2026-02-25 15:00:00
```

---

### lango cron delete

Delete a cron job by ID or name.

```
lango cron delete <id-or-name>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `id-or-name` | Yes | Job UUID or name |

**Example:**

```bash
$ lango cron delete daily-news
Cron job "daily-news" deleted.

$ lango cron delete a1b2c3d4-5678-...
Cron job "a1b2c3d4-5678-..." deleted.
```

---

### lango cron pause

Pause a running cron job. The job remains registered but will not execute until resumed.

```
lango cron pause <id-or-name>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `id-or-name` | Yes | Job UUID or name |

**Example:**

```bash
$ lango cron pause health-check
Cron job "health-check" paused.
```

---

### lango cron resume

Resume a paused cron job.

```
lango cron resume <id-or-name>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `id-or-name` | Yes | Job UUID or name |

**Example:**

```bash
$ lango cron resume health-check
Cron job "health-check" resumed.
```

---

### lango cron history

Show execution history for a specific job or all jobs.

```
lango cron history [id-or-name] [--limit N]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `id-or-name` | No | Filter by specific job (shows all if omitted) |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit`, `-n` | int | `20` | Maximum entries to show |

**Example:**

```bash
$ lango cron history
JOB           STATUS     STARTED               DURATION   RESULT
daily-news    completed  2026-02-20 09:00:00    3.2s       Top 5 tech stories: 1. AI breakthrough...
health-check  completed  2026-02-20 14:00:00    1.1s       All endpoints healthy
health-check  failed     2026-02-20 13:00:00    5.0s       ERR: timeout connecting to API server

$ lango cron history daily-news --limit 5
```

---

## Workflow Commands

Manage multi-step workflow pipelines defined in YAML files. Workflows must be enabled in configuration (`workflow.enabled = true`). See [Workflow Engine](../automation/workflows.md) for YAML format details.

```
lango workflow <subcommand>
```

### lango workflow run

Execute a workflow from a YAML definition file. If the workflow has a schedule, it registers with the server instead of executing immediately.

```
lango workflow run <file.flow.yaml> [--schedule <cron>]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `file.flow.yaml` | Yes | Path to the workflow YAML file |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--schedule` | string | | Cron schedule to register (overrides YAML) |

**Example:**

```bash
$ lango workflow run ./daily-report.flow.yaml
Workflow: Daily Report Pipeline
Steps:    3

Executing workflow...

Workflow completed: completed

--- Step: fetch-data ---
Retrieved 42 records from the database...

--- Step: analyze ---
Analysis complete. Key findings: ...

--- Step: report ---
Report generated and sent to #reports channel.
```

If the workflow includes a schedule or you override with `--schedule`, it registers for recurring execution:

```bash
$ lango workflow run ./report.flow.yaml --schedule "0 8 * * MON"
Workflow: Weekly Report
Steps:    3
Schedule: 0 8 * * MON

Workflow has a schedule. Register it with the running server:
  POST /api/workflow/register with the YAML content
```

---

### lango workflow list

List workflow runs.

```
lango workflow list [--limit N]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit`, `-n` | int | `20` | Maximum entries to show |

**Example:**

```bash
$ lango workflow list
ID        WORKFLOW              STATUS     STEPS  STARTED
a1b2c3d4  Daily Report Pipeline completed  3/3    2026-02-20 09:00:00
e5f6g7h8  Data Migration        running    2/5    2026-02-20 14:30:00
i9j0k1l2  Weekly Summary        failed     1/4    2026-02-19 08:00:00
```

---

### lango workflow status

Show detailed status for a specific workflow run, including per-step progress.

```
lango workflow status <run-id>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `run-id` | Yes | Workflow run ID |

**Example:**

```bash
$ lango workflow status a1b2c3d4
Run ID:    a1b2c3d4
Workflow:  Daily Report Pipeline
Status:    running
Progress:  2/3 steps

Steps:
  fetch-data            completed     agent=researcher
  analyze               completed     agent=planner
  generate-report       running       agent=executor
```

---

### lango workflow cancel

Cancel a running workflow.

```
lango workflow cancel <run-id>
```

| Argument | Required | Description |
|----------|----------|-------------|
| `run-id` | Yes | Workflow run ID to cancel |

**Example:**

```bash
$ lango workflow cancel e5f6g7h8
Workflow run e5f6g7h8 cancelled.
```

---

### lango workflow history

Show workflow execution history across all workflows.

```
lango workflow history [--limit N]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit`, `-n` | int | `20` | Maximum entries to show |

**Example:**

```bash
$ lango workflow history
ID        WORKFLOW              STATUS     STEPS
a1b2c3d4  Daily Report Pipeline completed  3/3
e5f6g7h8  Data Migration        cancelled  2/5
i9j0k1l2  Weekly Summary        failed     1/4
m3n4o5p6  Daily Report Pipeline completed  3/3
```
