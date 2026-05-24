## Why

The CLI core overview still describes inline tool approval interrupts as `d` deny, while the actual runtime and detailed feature docs expose `d/Esc` as the shared deny path. That leaves the first-touch CLI docs slightly behind the current interaction contract.

## What Changes

- Update the CLI core overview to use the unified `d/Esc` deny wording.
- Record the sync in OpenSpec.

## Impact

- Keeps first-touch CLI docs aligned with the actual approval surface.
- Removes a small but visible wording mismatch.
