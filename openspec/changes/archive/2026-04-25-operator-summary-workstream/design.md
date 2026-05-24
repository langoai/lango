## Design Summary

This workstream adds:

- `lango status dead-letter-summary`

Flow:

1. read the current dead-letter backlog through the existing list surface
2. aggregate a compact overview in the CLI layer
3. render the result as `table` or `json`

The first summary slice intentionally stays narrow:

- `total_dead_letters`
- `retryable_count`
- `by_adjudication`
- `by_latest_family`

The slice does not add a new backend summary contract or cockpit summary pane.
