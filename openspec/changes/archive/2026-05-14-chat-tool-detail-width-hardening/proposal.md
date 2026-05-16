## Why

Tool lifecycle rows already preserve param previews and output/error previews across state transitions, but the renderer still truncates only the raw strings before padding. On narrow terminals, the final indented detail line can still exceed the available width, and whitespace-heavy output can remain inconsistently normalized.

## What Changes

- Normalize tool preview and output text to single-line whitespace.
- Clamp the final rendered preview/output lines to the available width after padding.
- Clamp the header line to the available width as well.
- Add regressions for narrow-width tool detail rendering.
- Record the contract in OpenSpec and downstream cockpit docs.

## Impact

- Makes tool lifecycle rows match the same width-safety baseline as other compact transcript surfaces.
- Prevents output/error previews from destabilizing narrow chat layouts.
