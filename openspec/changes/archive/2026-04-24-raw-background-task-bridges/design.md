## Design Summary

This slice extends the post-adjudication detail view with a thin raw background-task bridge.

The canonical receipts-backed detail remains primary. On top of it, the service now optionally attaches:

- `latest_background_task`
  - `task_id`
  - `status`
  - `attempt_count`
  - `next_retry_at`

Selection rule:

- current transaction
- current adjudication outcome
- latest matching post-adjudication task using the existing retry key format

If there is no matching task, the bridge is `null`.

No new background stores, tools, or list-surface fields are introduced.
