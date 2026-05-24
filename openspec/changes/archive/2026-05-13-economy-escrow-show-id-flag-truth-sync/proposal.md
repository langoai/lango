## Why

`lango economy escrow show --id` now provides live-status guidance, but the CLI flag description still says "future use". That makes the command help internally inconsistent with the actual behavior and the updated docs.

## What Changes

- Replace the stale `--id` flag description in `lango economy escrow show`.
- Add a CLI regression that locks the updated help text.
- Sync the on-chain escrow spec to require truthful `--id` help wording.

## Impact

- `onchain-escrow`: CLI help stays consistent with the actual `--id` behavior.
- Operator UX: help output no longer advertises an implemented flag as future work.
