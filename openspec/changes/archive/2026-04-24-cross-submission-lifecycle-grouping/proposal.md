# Proposal

## Why

The dead-letter backlog already exposes current-submission retry lifecycle signals, but operators still need a simple transaction-wide view that includes historical submissions when triaging the current dead-letter row.

## What Changes

- extend the post-adjudication status read model with:
  - `transaction_global_total_retry_count`
  - `transaction_global_any_match_families`
- add filters for:
  - `transaction_global_total_retry_count_min`
  - `transaction_global_total_retry_count_max`
  - `transaction_global_any_match_family`
- update the public architecture docs and sync the OpenSpec requirements
- archive the completed change after sync
