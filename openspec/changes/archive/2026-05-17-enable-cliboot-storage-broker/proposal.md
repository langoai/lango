## Why

The bootstrap lifecycle spec requires normal application starts to use broker-owned database initialization. The shared CLI bootstrap helpers currently call `bootstrap.Run` without `StartStorageBroker`, so production entry points that use `cliboot.BootResult` can silently fall back to parent-owned SQLite handles.

That bypasses the storage ownership boundary that brokered storage is meant to enforce and leaves the most common runtime paths less isolated than doctor/settings/onboard, which already opt into broker startup.

## What Changes

- Update shared `cliboot` bootstrap helpers to request `StartStorageBroker: true`.
- Add regression tests around `cliboot.BootResult` and `cliboot.Config` so future callers cannot accidentally drop broker startup.
- Keep the lower-level `bootstrap.Run` option explicit for tests and specialized callers.

## Impact

- Production CLI/app entry points that use `cliboot` will enter the broker-owned DB path by default.
- Commands that only need config still close the bootstrap result after reading config.
- Existing bootstrap tests for explicit non-broker mode remain valid.
