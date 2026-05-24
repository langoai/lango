## Context

The shared prompt package already exposes `RequireInteractiveTerminal(message string)` and uses an injectable terminal check. Security commands rely on that helper for interactive-only flows. `onboard` and `settings` should reuse the same contract instead of adding another terminal detection mechanism.

## Design

Add a package-level `requireInteractiveTerminal` seam in both `internal/cli/onboard` and `internal/cli/settings`, initialized to `prompt.RequireInteractiveTerminal`. `NewCommand().RunE` calls that seam before `runOnboard` or `runSettings`.

The command-specific messages should be:

- `onboard requires an interactive terminal; use 'lango config create --preset <name>' or 'lango config import' for scripted setup`
- `settings requires an interactive terminal; use 'lango config import' or 'lango config set' for scripted configuration`

## Testing

Package tests override the seam to return a sentinel error and replace the run function seam, so the test proves the guard executes before bootstrap/TUI work. This keeps the test deterministic and avoids starting the real bootstrap stack.

## Non-Goals

- Do not add a non-interactive onboard wizard.
- Do not change `lango config` behavior.
- Do not alter Bubble Tea models or form navigation.
