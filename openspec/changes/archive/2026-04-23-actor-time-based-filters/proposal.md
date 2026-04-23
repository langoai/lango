# Proposal

## Why

The dead-letter backlog already supports filtering by adjudication, retry attempts, and receipt-ID query, but operators still need to narrow the backlog by who last requested manual replay and when the latest dead-letter happened.

## What Changes

- extend the post-adjudication status read model with latest manual replay actor and latest dead-letter timestamp
- add `manual_replay_actor`, `dead_lettered_after`, and `dead_lettered_before` filters to the backlog list
- surface the new actor/time fields through the existing read-only meta tool
- update the public architecture docs and sync the OpenSpec requirements
- archive the completed change after sync
