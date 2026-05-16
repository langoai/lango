# Librarian Commands

Commands for inspecting the proactive knowledge librarian. Output-capable commands accept only `table` or `json` for `--output`; unknown values fail fast before config or bootstrap loading begins.

## lango librarian status

Show the active librarian configuration from the current config profile.

```bash
lango librarian status [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango librarian status --output json
{
  "enabled": false,
  "observation_threshold": 2,
  "inquiry_cooldown_turns": 3,
  "max_pending_inquiries": 2,
  "auto_save_confidence": "high"
}
```

## lango librarian inquiries

List pending librarian inquiries from storage.

```bash
lango librarian inquiries [--limit <n>] [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | `20` | Maximum number of inquiries to show |
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango librarian inquiries --output json
[]
```
