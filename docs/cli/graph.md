# Graph CLI Reference

## Overview

The `lango graph` command group manages the knowledge graph store.

The graph store must be enabled in configuration (`graph.enabled`) and have a configured database path.
Most subcommands use `table` or `json` output and reject unknown `--output` values before loading graph state.

## Commands

### `lango graph status`

Show graph-store status.

```
lango graph status [--output table|json]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `table` | Output format: `table` or `json` |

### `lango graph query`

Query triples by subject, object, or subject+predicate.

```
lango graph query [--subject <value>] [--predicate <value>] [--object <value>] [--limit N] [--output table|json]
```

At least one of `--subject` or `--object` is required.
`--predicate` can only be used together with `--subject`.

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--subject` | `""` | Filter by subject |
| `--predicate` | `""` | Filter by predicate (requires `--subject`) |
| `--object` | `""` | Filter by object |
| `--limit` | `0` | Limit number of results (`0` = unlimited) |
| `--output` | `table` | Output format: `table` or `json` |

### `lango graph stats`

Show graph statistics.

```
lango graph stats [--output table|json]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `table` | Output format: `table` or `json` |

### `lango graph clear`

Clear all triples from the knowledge graph.

```
lango graph clear [--force]
```

Without `--force`, the CLI prints a confirmation prompt before deletion.

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--force` | `false` | Skip the confirmation prompt |

### `lango graph add`

Add a subject-predicate-object triple to the graph store.

```
lango graph add --subject <value> --predicate <value> --object <value> [--output table|json]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--subject` | required | Subject of the triple |
| `--predicate` | required | Predicate (relationship) of the triple |
| `--object` | required | Object of the triple |
| `--output` | `table` | Output format: `table` or `json` |

### `lango graph export`

Export all triples from the graph store.

```
lango graph export [--format json|csv]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--format` | `json` | Export format: `json` or `csv` |

### `lango graph import <file>`

Import triples from a JSON file.

```
lango graph import <file> [--output table|json]
```

The input file must contain a JSON array of triple objects.

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | `table` | Output format: `table` or `json` |

## Notes

- `status`, `query`, `stats`, `add`, and `import` write through the Cobra command output stream so wrappers can capture both table and JSON output directly.
- `export` supports `json` and `csv`; other values fail with an actionable error.
- `clear` uses a confirmation prompt unless `--force` is passed.
