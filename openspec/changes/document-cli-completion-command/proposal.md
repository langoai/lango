## Why

The root CLI exposes Cobra's generated `lango completion` command, but the public README and CLI reference do not mention it. This makes a shipped operator convenience feature effectively hidden and leaves the CLI documentation completeness guard unable to catch the drift.

## What Changes

- Document `lango completion` in the README top-level command list.
- Document `lango completion` in the CLI reference quick/core command tables, including the shell names exposed by Cobra.
- Extend documentation completeness guards so future public docs must keep the root completion command visible.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `cli-reference`: Public CLI reference coverage must include the root `lango completion` utility command when it is present in root help.

## Impact

- Public documentation: `README.md`, `docs/cli/index.md`
- Test guard: `internal/testutil`
- No runtime CLI behavior changes.
