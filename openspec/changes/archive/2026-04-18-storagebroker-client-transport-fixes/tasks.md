## 1. Transport Correctness

- [x] 1.1 Serialize broker request writes with a dedicated write mutex.
- [x] 1.2 Replace scanner-based stdout reading with a large-payload-safe decoder loop.
- [x] 1.3 Fix graceful shutdown so `Close()` attempts the shutdown RPC before closing the transport.

## 2. Verification

- [x] 2.1 Add transport regression tests for large responses and graceful shutdown.
- [x] 2.2 Run `go build ./...`, `go test ./...`, and `openspec validate --type change storagebroker-client-transport-fixes`.
