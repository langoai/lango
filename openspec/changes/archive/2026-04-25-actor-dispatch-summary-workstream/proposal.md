## Why

The richer dead-letter summary already surfaces dominant latest reasons, but operators still need to see who is most associated with current manual replay activity.

## What Changes

- extend `lango status dead-letter-summary`
- add `top_latest_manual_replay_actors`
- compute the top 5 latest manual replay actors from current backlog rows
- document the actor-summary slice in public docs and main OpenSpec specs

## Impact

- richer operator-facing dead-letter summary without a new backend service
- existing summary command stays stable and gains one additive actor section
- no new control-plane behavior
