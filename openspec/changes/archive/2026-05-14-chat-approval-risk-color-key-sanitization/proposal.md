## Why

The fullscreen approval dialog now sanitizes risk badge text, but badge color selection still keys off the raw `Risk.Level`. An escaped or malformed level can therefore render clean badge text while silently falling back to the wrong color mapping.

## What Changes

- Use the sanitized risk level when choosing the fullscreen approval badge color.
- Add regression coverage for escaped known risk levels.
- Record the sanitized risk-color key contract in OpenSpec and downstream cockpit docs.

## Impact

- Keeps fullscreen approval badge text and color mapping aligned under malformed input.
- Removes the last risk-badge mismatch between sanitized text and raw metadata.
