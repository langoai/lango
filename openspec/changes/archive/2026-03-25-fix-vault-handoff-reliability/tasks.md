## 1. Runtime cleanup

- [x] 1.1 Add streaming-path isolated child cleanup for iterator errors in `internal/adk/agent.go`.
- [x] 1.2 Add failed-turn dangling tool-call cleanup that closes unanswered parent-visible tool calls exactly once.
- [x] 1.3 Add regression tests covering isolated streaming failures and cleanup-safe retry history.

## 2. Structured recovery and tracing

- [x] 2.1 Carry failed specialist identity through `RecoveryContext`, `RecoveryEvent`, and reroute hints.
- [x] 2.2 Change structured recovery so post-specialist `ErrToolError` uses reroute recovery while pre-specialist retry behavior remains unchanged.
- [x] 2.3 Record structured recovery attempts in turn traces and switch trace writes to per-write detached timeout contexts.
- [x] 2.4 Add regression tests for specialist-aware recovery, recovery trace events, and long-turn trace persistence.

## 3. Downstream sync

- [x] 3.1 Update `README.md` and `docs/features/multi-agent.md` to describe reroute-aware structured recovery and trace-backed diagnostics.
- [x] 3.2 Run `go build ./...` and `go test ./...`, then sync specs and archive the OpenSpec change.
