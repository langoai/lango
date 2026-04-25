## Design Summary

This workstream keeps the existing summary command:

- `lango status dead-letter-summary`

and extends it with:

- `top_latest_manual_replay_actors`

Rules:

- aggregate from each row's current `latest_manual_replay_actor`
- include top 5 actors only
- each item contains `actor` and `count`
- existing summary fields remain unchanged
- the extension is additive in both `table` and `json`

The slice does not add dispatch breakdowns, grouped actor families, or a new summary backend path.
