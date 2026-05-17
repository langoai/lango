# Design: Background CLI Server Boundary

## Current Behavior

`internal/cli/bg.NewBgCmd` accepts a lazy `func() (*background.Manager, error)`. Package tests inject a real manager and the subcommands work in-process.

The root command cannot access the server process's in-memory manager. It wires the provider to an unconditional error. The existing error says to use `lango serve` first, which is incomplete because the CLI still lacks a remote client after the server starts.

## Proposed Approach

### Root Error

Replace the root provider error with a precise server-boundary message:

- Background task state is in-memory and owned by the running app/server process.
- This root CLI command is not yet connected to that process through a gateway API.
- Use in-app/cockpit task surfaces or agent `bg_*` tools for current task management.

This avoids the misleading "use `lango serve` first" implication.

### Command Help

Keep `internal/cli/bg` generic and manager-backed, but update the command long description to state that subcommands operate on the supplied in-process manager. Embedded callers and tests still work; root CLI callers receive the root boundary error.

### Docs and Guards

Update public command references to include the boundary caveat near the `lango bg` command table and the background automation page. Add a docs guard test that fails if public docs list `lango bg` commands without mentioning the root CLI's server-boundary limitation.

## Tradeoffs

This is intentionally a truth-alignment slice, not a remote-management implementation. Adding real HTTP-backed `lango bg` would require designing server routes, auth behavior, JSON schemas, and process-bound lifecycle semantics. That deserves a separate OpenSpec change because it changes the public API surface.
