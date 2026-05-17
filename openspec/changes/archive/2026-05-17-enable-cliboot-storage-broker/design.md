## Context

`internal/cli/cliboot.BootResult` and `Config` are the shared entry points used by serve, chat, cockpit, workbench, and most CLI commands. `bootstrap.Run` already supports broker startup through `Options.StartStorageBroker`, and several specialized commands set it directly.

The missing piece is the shared helper: it builds `bootstrap.Options{Version: Version}` only.

## Decision

Introduce a small package-level `bootstrapRun` variable in `internal/cli/cliboot` for tests and have both helpers call it with:

- `Version: Version`
- `StartStorageBroker: true`

This keeps the behavior centralized and avoids changing every command constructor.

## Test Strategy

Add `internal/cli/cliboot` tests that replace `bootstrapRun` with a capture function:

- `BootResult` passes `StartStorageBroker: true`.
- `Config` passes `StartStorageBroker: true` and still closes the returned result.

The tests verify option propagation without starting a real broker subprocess.
