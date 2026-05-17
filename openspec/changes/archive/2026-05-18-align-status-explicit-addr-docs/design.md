## Context

The runtime behavior for `lango status --addr` is already implemented: explicit addresses are normalized and the status output reports the same gateway target used for the `/health` probe. Public docs need to match that behavior, and a small docs guard should prevent future drift.

## Decisions

### Document The Override Contract

Extend the status command flags section to state both paths:

- omitted `--addr`: use configured `server.host` and `server.port`, then fallback to localhost
- supplied `--addr`: normalize the explicit address, probe it, and report the same normalized value in `gateway`

### Keep The Example Concrete

Update the custom gateway example and JSON schema sample so the visible `gateway` value is a custom address. This avoids teaching operators that output always reports localhost or config even when an override is supplied.

### Executable Guard

Extend `internal/testutil/gateway_cli_addr_docs_guard_test.go` with a status-specific assertion for the explicit-target wording. Keep the guard text-based because it verifies public docs, not runtime behavior. Runtime behavior remains covered by `internal/cli/status` and `internal/cli/clihttp` tests.

## Risks

- Text guards can become brittle if wording changes. Use short required phrases that encode the behavior without forcing a full paragraph.
- Do not broaden this change into a full CLI docs rewrite; the goal is downstream accuracy for the implemented status behavior.
