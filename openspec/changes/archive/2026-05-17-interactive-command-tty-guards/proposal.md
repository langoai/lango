## Why

`lango onboard` and `lango settings` are interactive Bubble Tea commands, but their package-level entrypoints start bootstrap and then run the TUI without an early interactive-terminal guard. Other interactive entrypoints such as `lango cockpit` and `lango chat` fail before startup work when stdin is not a terminal. Aligning onboard/settings with that behavior avoids confusing non-TTY runs in scripts, CI, and automation.

## What Changes

- Add command-level interactive-terminal guards to `lango onboard` and `lango settings`.
- Keep the guard package-local so direct package command usage is protected, not only the root command wiring.
- Add focused tests that prove both commands fail before bootstrap/TUI work when stdin is non-interactive.
- Update specs and public CLI docs to describe the non-interactive guidance.

## Impact

- Affected code: `internal/cli/onboard`, `internal/cli/settings`
- Affected tests: onboard/settings command tests
- Affected specs: `cli-onboard`, `cli-settings`, `test-coverage`
- Downstream docs: `docs/cli/core.md`
