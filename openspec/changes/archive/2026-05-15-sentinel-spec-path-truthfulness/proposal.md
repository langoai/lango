## Why

Sentinel and on-chain escrow specs still referenced deleted app-local tool-builder files even though sentinel and economy builders now live in their owning packages. That makes the specs materially stale and misleads anyone tracing production tool registration.

## What Changes

- sync sentinel and on-chain escrow specs to the current builder paths
- add an executable guard so the deleted builder-path claims cannot silently return

## Impact

- better alignment between specs and the current code layout
- less confusion about sentinel and economy tool ownership
- stronger regression protection for deleted builder-path drift
