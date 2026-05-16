## Why

The chat header and turn strip are measured as fixed-height single-line parts, but their renderers currently rely on `lipgloss.Width(...)` plus `Style.Width(...)` without truncating overflowing content first. On narrow terminals, those shell bars can wrap and break the fixed-height layout contract.

## What Changes

- Make the chat header and turn strip explicitly single-line and width-safe.
- Add regressions for narrow-width header/turn-strip rendering.
- Record the fixed-shell-bar contract in OpenSpec and downstream cockpit docs.

## Impact

- Preserves layout integrity for fixed shell parts on narrow terminals.
- Keeps header/turn strip behavior aligned with the rest of the chat width-hardening baseline.
