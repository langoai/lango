## 1. Shared Turn Runtime

- [x] 1.1 Introduce a shared `TurnRunner`/`TurnResult` abstraction for channel, gateway, and automation entrypoints.
- [x] 1.2 Add append-only per-turn trace creation, event recording, and final outcome classification.
- [x] 1.3 Migrate `internal/app/channels.go` and `internal/gateway/server.go` to the shared runner without changing external transport APIs.

## 2. Session Isolation Correctness

- [x] 2.1 Decouple child-session store activation from provenance/session-tree hooks.
- [x] 2.2 Prevent raw isolated specialist turns from being persisted into the parent session store under all runtime configurations.
- [x] 2.3 Persist classified discard/merge summaries for isolated runs without leaking raw child history.

## 3. Multi-Agent Completion And Recovery

- [x] 3.1 Enforce the specialist completion contract so tool-only terminal states become structured runtime outcomes.
- [x] 3.2 Replace same-tool churn detection with same-agent same-tool same-params containment and repeated-output awareness.
- [x] 3.3 Enforce evidence-only orchestrator recovery so the tool-less orchestrator never emits direct specialist tool calls.
- [x] 3.4 Add dedicated runtime error classifications and user-facing recovery messages for `empty_after_tool_use` and `loop_detected`.

## 4. Regression Harness

- [x] 4.1 Add sanitized transcript replay fixtures for the March 23 vault balance loop and related multi-agent failures.
- [x] 4.2 Add integration coverage for channel/gateway parity, isolation guarantees, loop containment, and trace-backed recovery.
- [x] 4.3 Expand shared test utilities as needed for replay-driven multi-agent runtime tests.

## 5. Downstream Surfaces And Verification

- [x] 5.1 Expose latest turn-trace summaries in operator diagnostics (CLI/TUI/doctor surface as selected in implementation).
- [x] 5.2 Update `README.md`, architecture/docs pages, prompts, and skills to match the new runtime guarantees and debugging workflow.
- [x] 5.3 Run `go build ./...` and `go test ./...` and fix all regressions.
- [x] 5.4 Run OpenSpec verify, sync affected specs, and archive the change after implementation.
