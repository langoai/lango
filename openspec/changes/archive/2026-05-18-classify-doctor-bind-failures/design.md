## Context

`NetworkCheck` formats the configured host and port through `gatewayaddr.ListenAddress`, then calls `net.Listen`. The current error path does not inspect the failure and always reports `Port <port> in use`.

## Decision

Classify listen failures at the doctor check boundary:

- Address-in-use errors remain port-conflict diagnostics.
- Other listen errors become bind-address diagnostics.
- The raw `net.Listen` error remains in `Result.Details` for troubleshooting.

## Alternatives Considered

- Parse platform-specific syscall errors only. Rejected because the code needs to remain portable and the existing tests run on developer machines with different operating systems.
- Add a new structured error type. Rejected because this is a CLI diagnostic boundary and no public Go API consumes the classification.
