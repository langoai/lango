## Why

The root `lango bg` command is visible in the CLI, but it currently fails with a boundary message because standalone CLI processes cannot access the running app/server process's in-memory background manager. That behavior is technically accurate but weak UX: users see a management command that cannot manage the tasks they created through live app, cockpit, or agent surfaces.

## What Changes

- expose a protected gateway REST surface for in-memory background task list/status/result/cancel operations
- keep background task state process-bound and ephemeral; the gateway only proxies the running process's manager
- refactor `internal/cli/bg` behind a small client interface with both in-process and gateway-backed implementations
- wire root `lango bg` to the gateway-backed client, using `--addr` or the configured server host/port
- preserve embedded/in-process bg command behavior for cockpit tests and internal callers
- update public docs and executable guards so `lango bg` is documented as gateway-backed remote management, not a standalone manager

## Impact

- `lango bg list/status/result/cancel` becomes usable against a running `lango serve`
- the command remains honest about ephemeral in-memory task state and server restart behavior
- new gateway endpoints expose task prompts/results, so they must use the existing gateway auth middleware when auth is configured
- existing tests that assert the root CLI boundary error must be replaced with remote-client behavior tests
