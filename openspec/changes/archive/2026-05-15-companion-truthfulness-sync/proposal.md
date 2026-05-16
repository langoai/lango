## Why

Public docs and the `companion-discovery` main spec still described an automatic mDNS-based companion discovery flow that the current runtime does not ship. The actual model is gateway-backed companion WebSocket connectivity through `/companion`.

## What Changes

- sync companion connectivity docs/specs to the current gateway-backed model
- add an executable guard so stale auto-discovery claims cannot silently return

## Impact

- docs and specs better reflect the shipped runtime
- less confusion around how companions actually connect
- stronger regression protection for stale discovery claims
