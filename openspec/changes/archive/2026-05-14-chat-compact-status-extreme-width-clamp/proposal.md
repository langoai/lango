## Why

Compact status and approval transcript rows now truncate their body text by width, but the final rendered line can still exceed the target width when the styled label and spacing already consume most of that budget. This shows up most clearly on very narrow terminals.

## What Changes

- Clamp the final rendered compact status/approval row to the requested width.
- Add regressions for very narrow width values.
- Record the extreme-width clamp contract in OpenSpec.

## Impact

- Fixes a real narrow-terminal layout bug.
- Keeps compact transcript rows within the viewport width even in extreme cases.
