## Why

Approval transcript events recently became richer with compact request-id and summary annotations. They still render through a width-unaware one-line block, so on narrow terminals the event row can overflow more easily than before.

## What Changes

- Make approval transcript event rendering width-aware and truncated for narrow terminals.
- Add regressions for the narrow-width path.
- Record the width-safety contract in OpenSpec.

## Impact

- Prevents approval event rows from overrunning narrow terminal layouts.
- Keeps the richer traceability copy while preserving compact rendering.
