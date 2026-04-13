# RunLedger Commands

!!! warning "Experimental"
    The RunLedger (Task OS) is experimental. See [RunLedger](../features/run-ledger.md) for feature details.

List, inspect, and manage durable execution runs powered by the RunLedger engine.

## lango run list

List recent runs.

```
lango run list [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--json` | `bool` | Output results as JSON |
| `--limit` | `int` | Maximum number of runs to list |

## lango run status

Show RunLedger configuration and current state.

```
lango run status [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--json` | `bool` | Output results as JSON |

## lango run journal

View the append-only event journal for a specific run.

```
lango run journal <run-id> [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--json` | `bool` | Output results as JSON |
| `--limit` | `int` | Maximum number of journal events to show |
