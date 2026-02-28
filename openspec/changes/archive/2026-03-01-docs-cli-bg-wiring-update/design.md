## Context

Recent commits added security hardening (keyring, SQLCipher, Cloud KMS), P2P session/sandbox management, and lifecycle management. The documentation (README.md, docs/cli/index.md, docs/architecture/project-structure.md, docs/index.md) has not been updated to reflect these changes. Additionally, `internal/cli/bg/` exists with full command implementation but is not wired in `cmd/lango/main.go`.

## Goals / Non-Goals

**Goals:**
- Wire `lango bg` command so it appears in `lango --help` output
- Update all documentation to accurately reflect current CLI commands and package structure
- Correct skills count/description to match actual state (scaffold only, no built-in skills)

**Non-Goals:**
- Implementing a gateway REST API for background tasks (bg commands use a stub provider)
- Adding new CLI functionality beyond wiring existing code
- Changing any internal package behavior

## Decisions

### bg command wiring uses a stub provider
The `background.Manager` is an in-memory component that only exists when the server is running. CLI commands cannot access it directly. The bg command is wired with a stub provider that returns an error directing users to `lango serve`. This matches the pattern used by other infrastructure commands.

**Alternative considered**: Gateway REST API — too much scope for a documentation update change.

### Documentation updates are additive
All changes add missing information rather than restructuring existing content. This minimizes diff size and review effort.

## Risks / Trade-offs

- [Risk] bg commands always error in standalone CLI mode → Users see clear error message directing them to start the server first. This is acceptable since the Manager is inherently server-scoped.
- [Risk] Documentation could drift again → Mitigated by the CLAUDE.md rule requiring downstream artifact updates with core code changes.
