## Why

`lango p2p connect` currently calls libp2p connect with `context.Background()`, so it is detached from Cobra command cancellation and any configured P2P timeout. A blocked network connect can make the operator CLI appear hung and delay cleanup of the ephemeral P2P node.

## What Changes

- Bound `lango p2p connect <multiaddr>` with a context derived from `cmd.Context()`.
- Use `p2p.handshakeTimeout` as the connect timeout, falling back to 30 seconds when unset or invalid.
- Ensure cleanup still runs when connect returns with timeout or cancellation.
- Return actionable timeout/cancellation errors that include the peer ID.
- Update public P2P CLI docs after the runtime behavior is verified.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `cli-p2p-management`: P2P connect must use bounded command-context-aware connection attempts.
- `downstream-docs-sync`: Public P2P CLI docs must describe the bounded connect behavior.
- `test-coverage`: P2P connect timeout, cancellation, and cleanup coverage must stay executable.

## Impact

- Affected code: `internal/cli/p2p/connect.go` and related tests.
- Affected docs: `docs/cli/p2p.md` and any quick reference that describes `lango p2p connect`.
- Runtime behavior changes only for `lango p2p connect` failure/cancellation paths.
