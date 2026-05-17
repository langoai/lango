## Context

`lango bg cancel <id>` is currently wired through Cobra and calls the background manager's `Cancel` method for the supplied in-process manager. The public automation docs correctly list the cancel command, but the section opener still says the CLI provides read-only management commands.

## Decision

Clarify the docs using two separate boundaries:

- `lango bg` commands are in-process management commands, not remote gateway operations for a running `lango serve` process.
- `list`, `status`, and `result` inspect task state, while `cancel` requests cancellation for a pending or running task when the command is supplied an in-process manager.

This keeps the previous server-boundary caveat while removing the inaccurate read-only claim.

## Alternatives Considered

- Removing `cancel` from the docs would hide an implemented command and violate the repository preference for additive feature growth.
- Calling the entire CLI mutating would be too broad because most subcommands are inspect-only.
