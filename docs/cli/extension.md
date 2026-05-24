# Extension CLI Reference

## Overview

The `lango extension` command group installs, inspects, lists, and removes extension packs.
Extension packs bundle skills, modes, and prompts.

The install flow is intentionally inspect-first:

- `inspect` prints a side-effect-free report
- `install` always prints the same inspect report before writing files
- `remove` prints the paths it will delete before confirmation

`inspect`, `install`, and `list` support `table`, `json`, and `plain` output modes.
When `--output` is omitted, the CLI auto-detects `table` on a TTY and `plain` otherwise.
Unknown `--output` values fail fast with an actionable error before any extension command continues.

Interactive `install` and `remove` confirmations require a TTY.
For scripted runs, pass `--yes`.

## Commands

### `lango extension inspect <source>`

Print a side-effect-free report about an extension pack.

```
lango extension inspect <source> [--output table|json|plain]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | auto-detected | Output format: `table`, `json`, or `plain` |

**Examples:**

```bash
# Inspect a local pack
lango extension inspect ./python-dev

# Inspect a git-backed pack at a specific ref
lango extension inspect https://example.com/pack.git#abc123 --output json
```

### `lango extension install <source>`

Install a pack with inspect + confirm.

```
lango extension install <source> [--yes] [--output table|json|plain]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--yes` | `false` | Skip the interactive confirmation; inspect output is still printed |
| `--output` | auto-detected | Output format for the inspect report: `table`, `json`, or `plain` |

**Examples:**

```bash
# Interactive install
lango extension install ./python-dev

# Scripted install
lango extension install --yes https://example.com/pack.git
```

### `lango extension list`

List installed extension packs.

```
lango extension list [--output table|json|plain]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--output` | auto-detected | Output format: `table`, `json`, or `plain` |

### `lango extension remove <name>`

Remove an installed pack.

```
lango extension remove <name> [--yes]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--yes` | `false` | Skip the interactive confirmation |

**Examples:**

```bash
# Interactive removal
lango extension remove python-dev

# Scripted removal
lango extension remove --yes python-dev
```

## Notes

- Local directory sources are detected from the filesystem; other values are treated as git sources.
- `install` and `remove` write confirmation and status messages through the Cobra command output stream, so wrappers can capture them directly.
- If stdin is not a TTY during an interactive run, the CLI tells you to pass `--yes` for scripted execution.
