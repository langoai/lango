## Design Summary

This slice extends:

- `lango status dead-letters`

with:

- `--latest-status-subtype`
- `--latest-status-subtype-family`

Validated values:

- subtype:
  - `retry-scheduled`
  - `manual-retry-requested`
  - `dead-lettered`
- family:
  - `retry`
  - `manual-retry`
  - `dead-letter`

The command keeps the existing:

- `table` default
- `json` support
- list/detail split
