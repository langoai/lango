## Why

The Approvals page now supports both `Tab` and `/` for section switching, but the help bar currently renders that binding as `tab//`, which is mechanically correct but visually awkward.

## What Changes

- Keep the dual-key section-toggle behavior.
- Clean up the help label to advertise the toggle keys in a readable format.
- Lock the rendered help contract in tests and spec text.

## Impact

- Operators see a clearer key hint without changing behavior.
- Runtime help, tests, and spec describe the same key label contract.
