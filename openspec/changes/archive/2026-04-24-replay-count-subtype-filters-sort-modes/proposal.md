# Proposal

## Why

The dead-letter backlog already supports actor/time filters plus reason/dispatch filters, but operators still need a way to narrow the list by subtype and manual replay count, and to reorder the backlog around the newest dead-letter or manual replay activity.

## What Changes

- extend the post-adjudication status read model with:
  - `latest_status_subtype`
  - `manual_retry_count`
  - `latest_manual_replay_at`
- add backlog filters for:
  - `latest_status_subtype`
  - `manual_retry_count_min`
  - `manual_retry_count_max`
- add `sort_by` with a bounded set of sort modes
- update the public architecture docs and sync the OpenSpec requirements
- archive the completed change after sync
