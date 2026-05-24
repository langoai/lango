## Why

The read model for post-adjudication dead letters is already landed, but operators still need to consume it through a real cockpit surface instead of raw tool outputs alone.

## What Changes

- add a read-only cockpit master-detail surface for dead-letter backlog triage
- reuse the existing list/detail status surfaces
- document the cockpit read surface in public docs and main OpenSpec specs

## Impact

- better operator usability
- no new backend endpoints
- no write actions in this slice
