## Why

Codex review against `main` identified two confirmed regressions in the `feature/lango-ontology` branch: (1) official release binaries built via Goreleaser lack the `vec` build tag, so vector/embedding features silently degrade to stubs; (2) running bare `lango` in headless/piped environments (CI, `</dev/null>`) crashes because bubbletea attempts to open `/dev/tty` without a TTY guard.

## What Changes

- Add `vec` build tag to both Goreleaser build configurations (`lango` and `lango-extended`) so release binaries include sqlite-vec embedding support
- Add TTY detection guard to root command, `cockpit`, and `chat` subcommands to gracefully handle non-interactive environments (help output for root, error for explicit TUI commands)

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `goreleaser-release`: Add `vec` build tag to both build targets for parity with Makefile/Dockerfile
- `cockpit-shell`: Add TTY guard before TUI startup to prevent crash in headless environments

## Impact

- `.goreleaser.yaml`: two build tag additions
- `cmd/lango/main.go`: TTY guard in 3 command RunE handlers, new import for `prompt` package
- Reuses existing `internal/cli/prompt.IsInteractive()` — no new dependencies
