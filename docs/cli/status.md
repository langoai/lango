---
title: Status Command
---

# lango status

Show a unified status dashboard combining health, configuration, and feature information.

## Synopsis

```bash
lango status [flags]
lango status dead-letter-summary [flags]
lango status dead-letters [flags]
lango status dead-letter <transaction-receipt-id> [flags]
lango status dead-letter retry <transaction-receipt-id> [flags]
```

## Description

The `status` command provides a single-screen overview of your Lango agent. It shows system info, active channels, and which features are enabled or disabled.

**Live mode**: When the gateway server is running, `status` probes the `/health` endpoint and reports whether the server is healthy.

**Config-only mode**: When the server is not running, `status` still shows configuration-based information (profile, provider, model, features, channels).

The `status` command also exposes dead-letter operator views:

- `lango status dead-letter-summary`
- `lango status dead-letters`
- `lango status dead-letter <transaction-receipt-id>`
- `lango status dead-letter retry <transaction-receipt-id>`

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `table` | Output format: `table` or `json` |
| `--addr` | `http://localhost:18789` | Gateway address to probe for live status |

## Dead-Letter Subcommands

### `lango status dead-letter-summary`

Show a global overview of the current dead-letter backlog.

The current summary surface includes:

- `total_dead_letters`
- `retryable_count`
- `top_limit`
- `by_adjudication`
- `by_latest_family`
- `by_reason_family`
  - grouped reason-family buckets derived from each backlog row's current `latest_dead_letter_reason`
  - initial taxonomy:
    - `retry-exhausted`
    - `policy-blocked`
    - `receipt-invalid`
    - `background-failed`
    - `unknown`
  - classification is case-insensitive and falls back to `unknown` when no built-in heuristic matches
- `by_actor_family`
  - grouped actor-family buckets derived from each backlog row's current `latest_manual_replay_actor`
  - initial taxonomy:
    - `operator`
    - `system`
    - `service`
    - `unknown`
  - classification is case-insensitive and falls back to `unknown` when no built-in heuristic matches
- `top_latest_dead_letter_reasons`
  - top `5` latest dead-letter reasons
  - each item includes:
    - `reason`
    - `count`
  - aggregated from each backlog row's current `latest_dead_letter_reason`
  - remains available alongside `by_reason_family` so operators can see both grouped families and raw latest reason strings
- `top_latest_manual_replay_actors`
  - top `5` latest manual replay actors
  - each item includes:
    - `actor`
    - `count`
  - aggregated from each backlog row's current `latest_manual_replay_actor`
  - remains available alongside `by_actor_family` so operators can see both grouped families and raw latest actor strings
- `by_dispatch_family`
  - grouped dispatch-family buckets derived from each backlog row's current `latest_dispatch_reference`
  - current classifier recognizes common leading tokens such as `dispatch`, `queue`, `worker`, `bridge`, `webhook`, and `unknown`
  - aliases such as `job`, `runner`, and `task` normalize to `worker`
  - otherwise the first normalized token is preserved as-is so unfamiliar prefixes still surface deterministically
- `top_latest_dispatch_references`
  - top latest dispatch references
  - each item includes:
    - `dispatch_reference`
    - `count`
  - aggregated from each backlog row's current `latest_dispatch_reference`
  - remains available alongside `by_dispatch_family` so operators can see both grouped families and raw latest dispatch strings
- `recent_dead_letter_trend`
  - a recent time-window summary over `latest_dead_lettered_at`
  - includes `window`, `bucket`, `windowed_count`, and per-bucket counts
  - rendered in table output as `Recent dead-letter trend`
  - rendered in JSON output as `recent_dead_letter_trend`

Flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `table` | Output format: `table` or `json` |
| `--top` | `5` | Top-N size for raw latest reason/actor/dispatch sections |
| `--trend-window` | `24h0m0s` | Time window for recent dead-letter trend output |
| `--trend-bucket` | `6h0m0s` | Bucket size for recent dead-letter trend output |

Examples:

```bash
lango status dead-letter-summary
lango status dead-letter-summary --output json
lango status dead-letter-summary --top 10
lango status dead-letter-summary --trend-window 72h --trend-bucket 24h
```

Table output includes `By reason family`, `By actor family`, `By dispatch family`, and `Recent dead-letter trend` sections while preserving the raw `Top Latest Dead-Letter Reasons`, `Top Latest Manual Replay Actors`, and `Top Latest Dispatch References` sections. JSON output includes `by_reason_family`, `by_actor_family`, `by_dispatch_family`, and `recent_dead_letter_trend`.

### `lango status dead-letters`

List the current dead-lettered post-adjudication backlog.

Flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `table` | Output format: `table` or `json` |
| `--query` | `""` | Substring filter over transaction/submission receipt IDs |
| `--adjudication` | `""` | Adjudication outcome filter: `release` or `refund` |
| `--latest-status-subtype` | `""` | Latest status subtype filter: `retry-scheduled`, `manual-retry-requested`, or `dead-lettered` |
| `--latest-status-subtype-family` | `""` | Latest status subtype family filter: `retry`, `manual-retry`, or `dead-letter` |
| `--any-match-family` | `""` | Any-match family filter: `retry`, `manual-retry`, or `dead-letter` |
| `--manual-replay-actor` | `""` | Latest manual replay actor filter |
| `--dead-lettered-after` | `""` | RFC3339 lower-bound timestamp filter for latest dead-letter time |
| `--dead-lettered-before` | `""` | RFC3339 upper-bound timestamp filter for latest dead-letter time |
| `--dead-letter-reason-query` | `""` | Latest dead-letter reason substring filter |
| `--latest-dispatch-reference` | `""` | Latest dispatch reference exact-match filter |

Examples:

```bash
lango status dead-letters
lango status dead-letters --query tx-123
lango status dead-letters --adjudication release --output json
lango status dead-letters --latest-status-subtype dead-lettered
lango status dead-letters --latest-status-subtype-family manual-retry
lango status dead-letters --any-match-family retry
lango status dead-letters --manual-replay-actor operator:alice
lango status dead-letters --dead-lettered-after 2026-04-25T09:00:00Z --dead-lettered-before 2026-04-25T18:00:00Z
lango status dead-letters --dead-letter-reason-query exhausted
lango status dead-letters --latest-dispatch-reference dispatch-7
```

### `lango status dead-letter <transaction-receipt-id>`

Show the current canonical dead-letter status for one transaction.

The output includes:

- canonical receipts-backed status
- latest retry / dead-letter summary
- `latest_background_task` when present

Flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `table` | Output format: `table` or `json` |

Examples:

```bash
lango status dead-letter tx-123
lango status dead-letter tx-123 --output json
```

### `lango status dead-letter retry <transaction-receipt-id>`

Request a retry for a dead-lettered post-adjudication execution.

Behavior:

- reads the current detail status first
- requires `can_retry=true`
- rejects before mutation when `can_retry=false`
- precheck rejection is surfaced as a retry-precheck error, not a mutation failure
- prompts for confirmation by default
- `--yes` skips the prompt
- reuses the existing retry control path
- success output means the retry request was accepted on the retry path, not that settlement execution already completed
- captures an immediate follow-up status snapshot after acceptance
- `--wait` keeps polling follow-up status until the observed follow-up changes or the wait timeout expires
- table output renders follow-up polling count, timeout/error state, and a structured follow-up block when available
- `json` output returns a structured retry-request result payload with `transaction_receipt_id`, `result`, `message`, `follow_up`, `follow_up_error`, `poll_count`, and `timed_out`
- `follow_up` includes `observed_at`, dead-letter/retryable booleans, latest subtype/family/reason/retry-attempt/dispatch data, and `background_task` when present
- invocation failures are surfaced separately as retry-request failures

Flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `table` | Output format: `table` or `json` |
| `--yes` | `false` | Skip the confirmation prompt |
| `--wait` | `false` | Poll follow-up status after retry request acceptance |
| `--wait-interval` | `2s` | Polling interval for retry follow-up status |
| `--wait-timeout` | `30s` | Polling timeout for retry follow-up status |

Examples:

```bash
lango status dead-letter retry tx-123
lango status dead-letter retry tx-123 --yes
lango status dead-letter retry tx-123 --yes --output json
lango status dead-letter retry tx-123 --yes --wait --wait-interval 1s --wait-timeout 30s
```

## Output Sections

### System

| Field | Description |
|-------|-------------|
| Server | `running` or `not running` (based on health probe) |
| Gateway | Configured host and port (e.g., `http://localhost:18789`) |
| Provider | AI provider and model (e.g., `openai (gpt-4o)`) |

### Channels

Lists all enabled messaging channels (telegram, discord, slack).

### Features

Shows each feature as enabled or disabled:

| Feature | Config Source |
|---------|-------------|
| Knowledge | `knowledge.enabled` |
| Embedding & RAG | `embedding.provider` (non-empty = enabled) |
| Graph | `graph.enabled` |
| Obs. Memory | `observationalMemory.enabled` |
| Librarian | `librarian.enabled` |
| Multi-Agent | `agent.multiAgent` |
| Cron | `cron.enabled` |
| Background | `background.enabled` |
| Workflow | `workflow.enabled` |
| MCP | `mcp.enabled` (with server count detail) |
| P2P | `p2p.enabled` |
| Payment | `payment.enabled` |
| Economy | `economy.enabled` |
| A2A | `a2a.enabled` |

## Examples

Full status dashboard (table format):

```bash
lango status
```

Machine-readable JSON output:

```bash
lango status --output json
```

Probe a custom gateway address:

```bash
lango status --addr http://192.168.1.10:18789
```

## JSON Schema

When using `--output json`, the response follows this structure:

```json
{
  "version": "1.2.3",
  "profile": "default",
  "serverUp": true,
  "gateway": "http://localhost:18789",
  "provider": "openai",
  "model": "gpt-4o",
  "features": [
    {
      "name": "Knowledge",
      "enabled": true
    },
    {
      "name": "MCP",
      "enabled": true,
      "detail": "2 server(s)"
    }
  ],
  "channels": ["telegram", "discord"],
  "serverInfo": {
    "healthy": true
  }
}
```

The `serverInfo` field is only present when the server is reachable. The `detail` field on features is optional and provides additional context (e.g., MCP server count).
