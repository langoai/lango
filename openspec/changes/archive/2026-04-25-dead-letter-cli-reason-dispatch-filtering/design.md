## Design Summary

This slice extends:

- `lango status dead-letters`

with:

- `--dead-letter-reason-query`
- `--latest-dispatch-reference`

Both values are passed through as strings without extra CLI validation.

The command keeps the existing:

- list/detail/retry split
- `table` default
- `json` support
