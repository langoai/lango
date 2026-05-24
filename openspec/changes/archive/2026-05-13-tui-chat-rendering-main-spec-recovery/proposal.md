## Why

Several archived chat-surface changes updated the same `Turn state strip` requirement, and the current main `tui-chat-rendering` spec now keeps only a small subset of the landed scenarios. That leaves the authoritative spec materially weaker than the implemented and tested chat approval/help surface.

## What Changes

- Restore the missing `Turn state strip` scenarios in the main `tui-chat-rendering` capability.
- Keep this change spec-only so the main spec once again matches the already-landed runtime, tests, and docs.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tui-chat-rendering`: Recover lost `Turn state strip` scenarios in the main capability spec.

## Impact

- Affected specs: `openspec/specs/tui-chat-rendering/spec.md`
- No runtime code changes
