## Context

`TestNetworkCheck_Run_PortAvailable` currently assumes `127.0.0.1:19999` is free. The suite already uses ephemeral listeners for occupied-port and IPv6 tests, so the availability test should follow the same pattern.

## Decision

Use a test helper that asks the OS for an ephemeral loopback TCP port, closes the listener, then passes that port into `NetworkCheck`.

## Tradeoffs

There remains a small race between releasing the helper listener and running the check, but this is materially safer than a fixed port and matches the existing IPv6 availability helper pattern.
