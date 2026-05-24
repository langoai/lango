## Design Summary

This workstream keeps the existing summary command:

- `lango status dead-letter-summary`

and extends it with:

- `top_latest_dead_letter_reasons`

Rules:

- aggregate from each row's current `latest_dead_letter_reason`
- include top 5 reasons only
- each item contains `reason` and `count`
- existing summary fields remain unchanged
- the extension is additive in both `table` and `json`

The slice does not add grouped reason families, actor/dispatch breakdowns, or a new summary backend path.
