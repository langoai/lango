## Why

The dead-letter operator surfaces already support backlog listing, per-transaction inspection, and retry, but operators still lack a fast overview of current backlog shape. A first summary surface is needed before richer grouped/operator dashboards.

## What Changes

- add `lango status dead-letter-summary`
- summarize the current dead-letter backlog using the existing list read model
- expose total dead letters, retryable count, adjudication buckets, and latest-family buckets
- document the landed summary surface in public docs and main OpenSpec specs

## Impact

- first operator-facing dead-letter summary surface
- no new backend summary service
- CLI overview stays aligned with the same dead-letter read model as the existing list/detail surfaces
