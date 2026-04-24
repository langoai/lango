# Proposal

## Why

The dead-letter backlog already supports manual retry count and raw subtype filtering, but operators still need a simpler lifecycle density view and a grouped family lens on the latest subtype.

## What Changes

- extend the post-adjudication status read model with:
  - `total_retry_count`
  - `latest_status_subtype_family`
- add backlog filters for:
  - `total_retry_count_min`
  - `total_retry_count_max`
  - `latest_status_subtype_family`
- update the public architecture docs and sync the OpenSpec requirements
- archive the completed change after sync
