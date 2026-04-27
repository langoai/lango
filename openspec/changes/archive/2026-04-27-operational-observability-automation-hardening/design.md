## Design Summary

This batch is intentionally narrow and does not reopen product semantics.

### Operational logging policy

For post-adjudication retry/dead-letter evidence writes and team-reputation bridge reputation/trust-entry failures:

- keep the execution path best-effort
- raise the log level to `Errorw` for lost canonical evidence or lost trust-entry updates
- use `Warnw` for downstream kick operations that fail after policy evaluation

This keeps behavior non-destructive while making the failure visible.

### CLI automation hardening

The dead-letter CLI surface now exposes:

- `--offset`
- `--limit`
- `--actor`

The existing retry bridge still injects a stable local fallback principal, but explicit actor override happens one layer earlier in the command path.

### JSON error contract

Dead-letter status subcommands in JSON mode now return:

```json
{
  "result": "error",
  "error": "..."
}
```

instead of plain text. This applies only when `--output json` is selected.
