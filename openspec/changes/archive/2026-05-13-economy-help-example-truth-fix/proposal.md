## Why

The top-level `lango economy` help text still shows example commands for subcommands that do not exist in the current CLI surface (`risk assess`, `pricing quote`, `negotiate list`, `escrow status --escrow-id`). That makes the very first operator guidance for the economy command group inaccurate.

## What Changes

- Replace the stale `lango economy` help examples with commands that actually exist in the current CLI surface.
- Add a CLI regression that locks the corrected help output.
- Sync the economy CLI spec so help examples are expected to stay on the real status/show surface.

## Impact

- `economy-cli`: top-level command help matches the actual subcommand tree.
- Operator UX: users are no longer shown nonexistent example commands in the first economy help screen.
