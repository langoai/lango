## Context

Codex review identified two regressions in the `feature/lango-ontology` branch vs `main`:

1. **Goreleaser vec tag**: Makefile and Dockerfile both use `-tags "fts5,vec"`, but `.goreleaser.yaml` only specifies `fts5`. Release binaries therefore compile the stub (`sqlite_vec_stub.go`) instead of the real sqlite-vec implementation.
2. **TTY guard**: Root command (`lango`) unconditionally calls `runCockpit()`, which starts bubbletea with `WithAltScreen()`. In non-TTY environments, bubbletea opens `/dev/tty` and fails with "could not open a new TTY".

Note: A third Codex finding (Dockerfile `fts5` vs `sqlite_fts5`) was verified as a false positive — mattn/go-sqlite3 v1.14.33 accepts both tags (`//go:build sqlite_fts5 || fts5`).

## Goals / Non-Goals

**Goals:**
- Release binaries include sqlite-vec support (parity with local/Docker builds)
- Bare `lango` in headless environments outputs help instead of crashing
- Explicit TUI commands (`cockpit`, `chat`) return a clear error in non-TTY environments

**Non-Goals:**
- Adding vec/fts5 tags to CI workflow (separate concern, may need CGO dependency setup)
- Implementing a non-interactive mode for the agent (out of scope)

## Decisions

1. **Add `vec` to both Goreleaser builds** — straightforward config alignment with Makefile/Dockerfile. No alternative considered; this is a missing-config bug.

2. **Guard at command level, not inside `runCockpit`/`runChat`** — placing the TTY check in cobra `RunE` avoids expensive bootstrap (DB open, logging init) before failing. The existing `prompt.IsInteractive()` function (`internal/cli/prompt/prompt.go:14`) is reused.

3. **Root command → `cmd.Help()` (exit 0); explicit subcommands → error** — when a user runs bare `lango` without TTY, showing help is most useful for discoverability. When they explicitly request `cockpit` or `chat`, an error is appropriate since they intended TUI mode.

## Risks / Trade-offs

- **[Low] Goreleaser CGO dependency**: The `vec` tag requires sqlite-vec CGO bindings at build time. Goreleaser already sets `CGO_ENABLED=1` and the dependency is in `go.mod`, so no additional risk.
- **[Low] IsInteractive false negatives**: Some terminal emulators or SSH sessions might report non-interactive incorrectly. Mitigation: `golang.org/x/term.IsTerminal` is the standard Go approach and is already used in 8+ places in this codebase.
