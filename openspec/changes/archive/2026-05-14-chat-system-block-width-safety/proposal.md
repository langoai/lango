## Why

Recent hardening focused on compact status and approval rows, but the `System` transcript block still renders raw body text without any width-aware clamping. On narrow terminals, long system notes can overflow the viewport.

## What Changes

- Make system transcript blocks width-aware.
- Add regressions for narrow-width system rendering.
- Record the width-safety contract in OpenSpec.

## Impact

- Prevents narrow-terminal overflow for system transcript entries.
- Brings another transcript item kind into the same production-hardening baseline.
