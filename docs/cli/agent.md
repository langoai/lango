# Agent CLI Reference

## Overview

The `lango agent` command group covers two surfaces:

- inspection commands such as `status`, `list`, `tools`, and `hooks`
- diagnostics commands such as `trace`, `graph`, and trace-derived metrics

Inspection-oriented commands are documented in [Agent & Memory](agent-memory.md).
This page focuses on the diagnostics surface that uses bootstrap-backed trace storage.

All diagnostics commands accept `--output table|json` and reject unknown values before bootstrap-dependent work begins.
Human-readable and JSON output are routed through the Cobra command writer so wrappers and test harnesses can capture them directly.

## Commands

### `lango agent trace list`

List recent turn traces.

```
lango agent trace list [--session <key>] [--outcome <outcome>] [--limit N] [--output table|json]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--session` | `""` | Filter by session key |
| `--outcome` | `""` | Filter by trace outcome |
| `--limit` | `20` | Maximum traces to return |
| `--output` | `table` | Output format: `table` or `json` |

### `lango agent trace show <trace-id>`

Show a detailed event timeline for one trace.

```
lango agent trace show <trace-id> [--output table|json]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `table` | Output format: `table` or `json` |

### `lango agent trace metrics`

Show trace-derived per-agent performance metrics.
This is distinct from `lango metrics agents`, which reports token usage rather than trace outcomes and durations.

```
lango agent trace metrics [--agent <name>] [--output table|json]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--agent` | `""` | Filter by agent name |
| `--output` | `table` | Output format: `table` or `json` |

### `lango agent graph <session-key>`

Show the delegation graph for a session.

```
lango agent graph <session-key> [--output table|json]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `table` | Output format: `table` or `json` |

## Notes

- `trace`, `graph`, and trace-derived metrics use bootstrap-backed storage rather than config-only inspection.
- `lango agent trace metrics` lives under the `trace` command tree; it is not exposed as a root-level `lango agent metrics` command.
- If no diagnostics data exists, the commands render explicit empty-state output instead of failing silently.
