## 1. Payment Persistence Interfaces

- [x] 1.1 Add explicit payment transaction-store and usage-reader interfaces with Ent-backed adapters
- [x] 1.2 Update payment service and settlement service to depend on explicit storage-facing persistence interfaces

## 2. Wiring Migration

- [x] 2.1 Rewire payment CLI dependency initialization to use storage-facing payment capabilities only
- [x] 2.2 Rewire app payment and P2P settlement wiring to use storage-facing payment capabilities only

## 3. Verification

- [x] 3.1 Add or update tests covering payment send, settlement status updates, and limiter usage through the new interfaces
- [x] 3.2 Run `go build ./...`, `go test ./...`, and `openspec validate --type change broker-ownership-mutators`
