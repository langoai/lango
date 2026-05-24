# Proposal: Background CLI Server Boundary

## Summary

Make `lango bg` truthful about its current runtime boundary. Background tasks live in an in-memory manager owned by the running app/server process, while the root CLI currently has no HTTP-backed client path into that manager. The CLI, help text, and public docs should clearly explain that boundary instead of presenting `lango bg list/status/cancel/result` as ordinary standalone commands.

## Motivation

Public docs list `lango bg list`, `lango bg status`, `lango bg cancel`, and `lango bg result` as runnable operator commands. In the root CLI, however, `cmd/lango/main.go` wires `lango bg` to a provider that always returns `bg commands require a running server (use 'lango serve' first)`. Even if `lango serve` is running, the current command has no gateway API client to inspect that process's in-memory background manager.

That mismatch creates a production-readiness problem: users are instructed to run commands that cannot work from the standalone CLI process. Until a real remote management API exists, the product should be explicit that these subcommands only work when embedded with an in-process manager and that root CLI management is not yet connected to the running server.

## Scope

- Keep the existing in-process `internal/cli/bg` manager-backed command behavior for tests and embedded callers.
- Improve the root CLI failure copy so it states the actual boundary and does not imply that simply starting `lango serve` makes the current command work.
- Update `lango bg --help` / command long description to describe in-process manager scope.
- Update public docs that list `lango bg` commands with a server-bound/in-process caveat.
- Add executable guards so README/docs cannot list `lango bg` commands without the caveat.

## Non-Goals

- Add new gateway HTTP endpoints for background task management.
- Persist background task state across processes.
- Add `bg submit`; task submission remains through agent tools.
