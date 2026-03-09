# Metrics Commands

Commands for viewing observability metrics including token usage, tool execution stats, and agent performance. Requires a running `lango serve` instance.

```
lango metrics [subcommand] [flags]
```

### Persistent Flags

All metrics commands share these flags:

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format: `table` or `json` |
| `--addr` | string | `http://localhost:18789` | Gateway address |

---

## lango metrics

Show a system metrics snapshot summary including uptime, total token usage, and tool execution count.

```
lango metrics [--output table|json] [--addr <url>]
```

**Example:**

```bash
$ lango metrics
=== System Metrics ===

Uptime:           2h15m30s
Total Input:      145200 tokens
Total Output:     52800 tokens
Tool Executions:  342

$ lango metrics --output json
{
  "uptime": "2h15m30s",
  "tokenUsage": {
    "inputTokens": 145200,
    "outputTokens": 52800
  },
  "toolExecutions": 342
}
```

---

## lango metrics sessions

Show per-session token usage breakdown including input/output tokens and request count.

```
lango metrics sessions [--output table|json] [--addr <url>]
```

**Example:**

```bash
$ lango metrics sessions
SESSION                   INPUT   OUTPUT  TOTAL    REQUESTS
abc123def456ghij78901...  45200   12800   58000    24
xyz789abc012defg34567...  32000   9400    41400    18

$ lango metrics sessions --output json
```

---

## lango metrics tools

Show per-tool execution statistics including call count, errors, error rate, and average duration.

```
lango metrics tools [--output table|json] [--addr <url>]
```

**Example:**

```bash
$ lango metrics tools
TOOL              COUNT  ERRORS  ERROR RATE  AVG DURATION
web_search        85     2       2.4%        1.2s
code_review       42     0       0.0%        3.5s
file_read         156    1       0.6%        0.1s
memory_store      63     0       0.0%        0.2s
```

---

## lango metrics agents

Show per-agent token usage breakdown including input/output tokens and tool call count.

```
lango metrics agents [--output table|json] [--addr <url>]
```

**Example:**

```bash
$ lango metrics agents
AGENT       INPUT   OUTPUT  TOOL CALLS
executor    82000   31200   198
researcher  45200   15600   96
planner     18000   6000    48
```

---

## lango metrics history

Show historical token usage from the database for the specified number of days.

```
lango metrics history [--days <n>] [--output table|json] [--addr <url>]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--days` | int | `7` | Number of days to query |

**Example:**

```bash
$ lango metrics history --days 3
Token usage history (last 3 days)
Records: 156 | Total Input: 520000 | Total Output: 185000

TIME              PROVIDER  MODEL               INPUT   OUTPUT
2026-03-07 14:30  openai    gpt-4o              4200    1800
2026-03-07 14:25  anthropic claude-sonnet-4-6... 3800    1200
2026-03-07 13:50  openai    gpt-4o              5100    2400

$ lango metrics history --days 7 --output json
```
