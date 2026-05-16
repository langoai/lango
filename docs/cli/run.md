# RunLedger Commands

!!! warning "Experimental"
    The RunLedger (Task OS) is experimental. See [RunLedger](../features/run-ledger.md) for feature details.

List, inspect, and manage durable execution runs powered by the RunLedger engine.
All three subcommands write through the Cobra command output stream so wrappers and test harnesses can capture RunLedger inspection output directly.

## lango run list

List recent runs.

```
lango run list [--output table|json] [--limit <n>]
```

| Flag | Type | Description |
|------|------|-------------|
| `--output` | `string` | Output format (`table` or `json`) |
| `--limit` | `int` | Maximum number of runs to list (`0` uses config default) |

## lango run status

Show RunLedger configuration and current state.

```
lango run status [--output table|json]
```

| Flag | Type | Description |
|------|------|-------------|
| `--output` | `string` | Output format (`table` or `json`) |

## lango run journal

View the append-only event journal for a specific run.

```
lango run journal <run-id> [--output table|json] [--limit <n>]
```

| Flag | Type | Description |
|------|------|-------------|
| `--output` | `string` | Output format (`table` or `json`) |
| `--limit` | `int` | Maximum number of journal events to show (`0` means all) |
