# Proposal

## Why

The dead-letter backlog already exposes latest subtype family and any-match families, but operators still need a single dominant family summary for the current submission lifecycle and a simple exact-match filter on that summary.

## What Changes

- extend the post-adjudication status read model with:
  - `dominant_family`
- add backlog filter:
  - `dominant_family`
- update the public architecture docs and sync the OpenSpec requirements
- archive the completed change after sync
