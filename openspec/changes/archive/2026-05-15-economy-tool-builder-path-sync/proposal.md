## Why

Some economy-related specs still referred to deleted app-local tool builder files even though sentinel and economy builders now live in their owning packages. Those stale paths make the specs materially misleading for anyone trying to trace the production registration path.

## What Changes

- sync sentinel/on-chain escrow specs to the current builder paths
- add an executable guard so the deleted app-local builder paths cannot silently return

## Impact

- specs better match the current production registration flow
- less confusion when tracing sentinel and economy tool ownership
- stronger regression protection for deleted builder-path drift
