## Overview

The root cause is that the CLI connect path ignores the command context and calls `Host().Connect(context.Background(), ...)`. This bypasses cancellation from callers and provides no CLI-level upper bound for peer connection attempts.

## Design

### Connect Context

Thread `cmd.Context()` into the connect implementation and derive a timeout child context from it:

- Use `boot.Config.P2P.HandshakeTimeout` when it is positive.
- Use 30 seconds when the configured timeout is zero or negative.
- Defer the timeout cancellation immediately after creating the child context.

### Testability

Keep the public command behavior thin by injecting the command context into the package-level connect seam. Add a lower-level host-connect seam so tests can observe the context without starting a real libp2p host.

### Error Handling

Wrap timeout and cancellation failures with peer ID context so users can see which peer failed:

- Timeout: `connect to <peer-id> timed out after <duration>: ...`
- Parent command deadline: `connect to <peer-id> timed out by command context deadline: ...`
- Cancellation: `connect to <peer-id> canceled: ...`
- Other connect errors keep the existing `connect to <peer-id>: ...` shape.

Cleanup must run for parse failures and connect failures exactly as it does today.

## Non-Goals

- Add retry/backoff.
- Add a new CLI flag for timeout.
- Change P2P node startup behavior outside `connect`.
