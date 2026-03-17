---
title: Status Command
---

# lango status

Show a unified status dashboard combining health, configuration, and feature information.

## Synopsis

```bash
lango status [flags]
```

## Description

The `status` command provides a single-screen overview of your Lango agent. It shows system info, active channels, and which features are enabled or disabled.

**Live mode**: When the gateway server is running, `status` probes the `/health` endpoint and reports whether the server is healthy.

**Config-only mode**: When the server is not running, `status` still shows configuration-based information (profile, provider, model, features, channels).

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `table` | Output format: `table` or `json` |
| `--addr` | `http://localhost:18789` | Gateway address to probe for live status |

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
