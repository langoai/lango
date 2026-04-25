## Design Summary

This workstream keeps the existing summary command:

- `lango status dead-letter-summary`

and extends it with:

- `top_latest_dispatch_references`

Rules:

- aggregate from each row's current `latest_dispatch_reference`
- include top 5 dispatch references only
- each item contains `dispatch_reference` and `count`
- existing summary fields remain unchanged
- the extension is additive in both `table` and `json`

The slice does not add grouped dispatch families, trend/time-window views, or a new summary backend path.
