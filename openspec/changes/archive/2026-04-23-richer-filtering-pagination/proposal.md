# Proposal

## Why

The first dead-letter browsing / status observation slice gave operators a read-only backlog and a per-transaction detail view, but the backlog still lacked the minimum filtering and pagination needed for practical triage.

## What Changes

- extend the post-adjudication status service with richer dead-letter list filters
- add `offset` / `limit` pagination metadata to the backlog response
- add lightweight navigation hints to the per-transaction detail view
- update the public architecture docs to describe the richer list/detail surface
- sync the OpenSpec requirements for the upgraded status tools and docs
- archive the completed change after sync
