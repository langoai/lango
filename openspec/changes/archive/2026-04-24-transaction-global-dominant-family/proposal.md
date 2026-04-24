# Proposal

## Why

The dead-letter backlog already exposes transaction-global total retry count and any-match family grouping, but operators still need a single transaction-global dominant family signal for quicker triage.

## What Changes

- extend the post-adjudication status read model with:
  - `transaction_global_dominant_family`
- add backlog filter:
  - `transaction_global_dominant_family`
- update the public architecture docs and sync the OpenSpec requirements
- archive the completed change after sync
