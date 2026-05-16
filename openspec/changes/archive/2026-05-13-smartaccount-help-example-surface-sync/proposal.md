## Why

The top-level `lango account` help text does not show the paymaster command family even though paymaster status and approval are part of the current smart-account CLI surface. That makes the first operator-facing overview less complete than the actual subcommand tree.

## What Changes

- Add a representative paymaster example to the top-level `lango account` help text.
- Add a CLI regression locking the updated help output.
- Sync the smart-account CLI docs and downstream spec to the same top-level help contract.

## Impact

- `smartaccount-downstream`: top-level CLI help better reflects the actual smart-account surface.
- Operator UX: users see paymaster operations immediately in the first help screen.
