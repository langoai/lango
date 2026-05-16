## Why

The approval-state turn strip is documented as following the shared `d/Esc` deny-key contract, but the runtime hint still renders `d(Esc)`. The behavior is correct; only the visible wording lags behind the rest of the approval surfaces.

## What Changes

- Align the approval-state turn strip hint with the shared `d/Esc` deny wording.
- Add regression coverage for the turn-strip copy.
- Record the explicit turn-strip wording contract in OpenSpec.

## Impact

- Removes the last visible deny-key wording mismatch inside the chat shell.
- Keeps turn strip, help bar, and public docs on the same operator-facing contract.
