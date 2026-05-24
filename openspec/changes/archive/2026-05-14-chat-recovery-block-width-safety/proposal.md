## Why

`renderRecoveryBlock()` currently ignores its `width` argument and emits a compact recovery row at natural length. That leaves recovery events exposed to narrow-terminal overflow even though neighboring compact transcript rows have already been hardened.

## What Changes

- Make recovery transcript rows width-aware.
- Add regressions for narrow-width recovery rendering.
- Record the width-safety contract in OpenSpec and public cockpit docs.

## Impact

- Prevents compact recovery rows from overrunning narrow terminals.
- Brings another transcript event surface into the same rendering-safety baseline.
