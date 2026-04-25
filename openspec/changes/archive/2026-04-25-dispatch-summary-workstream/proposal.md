## Why

The dead-letter summary surface already shows dominant latest reasons and actors, but operators still need to see which dispatch references are most associated with the current backlog.

## What Changes

- extend `lango status dead-letter-summary`
- add `top_latest_dispatch_references`
- compute the top 5 latest dispatch references from current backlog rows
- document the dispatch-summary slice in public docs and main OpenSpec specs

## Impact

- richer operator-facing dead-letter summary without a new backend service
- existing summary command stays stable and gains one additive dispatch section
- no new control-plane behavior
