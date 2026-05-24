# Design

The cron domain already stores per-job timeouts and the scheduler already
honors `Job.Timeout`. The CLI should expose that existing behavior rather than
introducing a new storage or scheduler path.

Implementation approach:

- Add `--timeout` to `lango cron add` and parse it with `time.ParseDuration`.
- Add `--deliver-to` as an alias for the existing `--deliver` string-slice
  flag by binding both flags to the same destination.
- Add optional `--id` flags to `delete`, `pause`, `resume`, and `history`.
- Resolve job selectors from either positional argument or `--id`; reject calls
  that provide neither or both to avoid ambiguous user intent.

Existing positional `id-or-name` arguments remain supported for backward
compatibility and for the shorter README examples.
