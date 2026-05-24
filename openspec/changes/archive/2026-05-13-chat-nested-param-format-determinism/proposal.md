## Why

The chat approval surfaces and tool previews now sort top-level param keys, but nested structured values still fall back to `fmt.Sprintf`. For map-backed nested payloads, that can reintroduce non-deterministic rendering even when the outer order is fixed.

## What Changes

- Format structured param values through stable JSON-style rendering instead of raw `%v` formatting.
- Reuse the same deterministic formatting across approval surfaces and tool previews.
- Add regressions for nested payload ordering and record the contract in OpenSpec.

## Impact

- Improves deterministic rendering for complex tool payloads.
- Makes approval and tool transcript surfaces more reproducible for operators and tests.
