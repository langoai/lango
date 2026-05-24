# Approval Commands

Commands for inspecting approval and interception policy. Output-capable commands accept only `table` or `json` for `--output`; unknown values fail fast before config loading begins.

## lango approval status

Show the current approval and security-interceptor configuration.

```bash
lango approval status [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango approval status --output json
{
  "interceptor_enabled": false,
  "approval_policy": "dangerous",
  "headless_auto_approve": false,
  "approval_timeout_sec": 0,
  "redact_pii": false
}
```
