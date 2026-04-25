## Why

The first dead-letter summary surface already shows backlog size and family distribution, but operators still need a quicker read on the dominant current failure reasons in the backlog.

## What Changes

- extend `lango status dead-letter-summary`
- add `top_latest_dead_letter_reasons`
- compute the top 5 latest dead-letter reasons from current backlog rows
- document the richer summary slice in public docs and main OpenSpec specs

## Impact

- richer operator-facing dead-letter overview without a new backend service
- existing summary command stays stable and gains one additive section
- no new control-plane behavior
