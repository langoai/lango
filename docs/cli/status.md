---
title: Status Command
---

# lango status

Show a unified status dashboard combining health, configuration, and feature information.

## Synopsis

```bash
lango status [flags]
lango status dead-letters [flags]
lango status dead-letter <transaction-receipt-id> [flags]
lango status dead-letter retry <transaction-receipt-id> [flags]
```

## Description

The `status` command provides a single-screen overview of your Lango agent. It shows system info, active channels, and which features are enabled or disabled.

**Live mode**: When the gateway server is running, `status` probes the `/health` endpoint and reports whether the server is healthy.

**Config-only mode**: When the server is not running, `status` still shows configuration-based information (profile, provider, model, features, channels).

The `status` command also exposes dead-letter operator views:

- `lango status dead-letters`
- `lango status dead-letter <transaction-receipt-id>`
- `lango status dead-letter retry <transaction-receipt-id>`

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `table` | Output format: `table` or `json` |
| `--addr` | `http://localhost:18789` | Gateway address to probe for live status |

## Dead-Letter Subcommands

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
| `--manual-replay-actor` | `""` | Latest manual replay actor filter |
| `--dead-lettered-after` | `""` | RFC3339 lower-bound timestamp filter for latest dead-letter time |
| `--dead-lettered-before` | `""` | RFC3339 upper-bound timestamp filter for latest dead-letter time |

Examples:

```bash
lango status dead-letters
lango status dead-letters --query tx-123
lango status dead-letters --adjudication release --output json
lango status dead-letters --latest-status-subtype dead-lettered
lango status dead-letters --latest-status-subtype-family manual-retry
lango status dead-letters --manual-replay-actor operator:alice
lango status dead-letters --dead-lettered-after 2026-04-25T09:00:00Z --dead-lettered-before 2026-04-25T18:00:00Z
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
- prompts for confirmation by default
- `--yes` skips the prompt
- reuses the existing retry control path

Flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `table` | Output format: `table` or `json` |
| `--yes` | `false` | Skip the confirmation prompt |

Examples:

```bash
lango status dead-letter retry tx-123
lango status dead-letter retry tx-123 --yes
lango status dead-letter retry tx-123 --yes --output json
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
