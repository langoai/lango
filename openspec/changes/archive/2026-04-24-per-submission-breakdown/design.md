## Design Summary

This slice extends the existing dead-letter backlog read model with a compact per-submission breakdown.

Each backlog row now includes `submission_breakdown`, ordered `oldest -> newest`, with one item per submission in the transaction:

- `submission_receipt_id`
- `retry_count`
- `any_match_families`

The computation stays read-only and on-read:

- scan all submissions attached to the current transaction
- summarize relevant `post_adjudication_retry` events per submission
- project the compact breakdown into the backlog row

No canonical transaction or submission state changes are introduced.
