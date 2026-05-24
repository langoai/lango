## Design Summary

This slice extends:

- `lango status dead-letters`

with:

- `--manual-replay-actor`
- `--dead-lettered-after`
- `--dead-lettered-before`

The time flags are validated as RFC3339 in the CLI before bridge invocation.

The command keeps the existing:

- list/detail/retry split
- `table` default
- `json` support
