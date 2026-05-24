# Proposal

## Why

The dead-letter backlog already exposes the family of the latest retry/dead-letter subtype, but operators still need a lightweight way to see every family touched by the current submission lifecycle and filter the backlog by one family at a time.

## What Changes

- extend the post-adjudication status read model with:
  - `any_match_families`
- add backlog filter:
  - `any_match_family`
- update the public architecture docs and sync the OpenSpec requirements
- archive the completed change after sync
