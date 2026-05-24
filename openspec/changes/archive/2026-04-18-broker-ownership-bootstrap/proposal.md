## Why

`StartStorageBroker=true` still opens the same SQLite database in both the broker and parent process. That means the broker does not actually own database initialization or bootstrap-time storage access.

## What Changes

- Move bootstrap config/profile reads and session bootstrap access onto broker-backed storage adapters.
- Add broker RPC support for config profile CRUD and session bootstrap lifecycle methods.
- Prepare bootstrap for a follow-up step where parent direct DB open can be removed safely.

## Capabilities

### Modified Capabilities
- `brokered-storage`: broker gains bootstrap-time config/profile/session ownership APIs.
- `session-store`: broker-backed session access becomes available for bootstrap/runtime wiring.

## Impact

- Affected code: `internal/storagebroker`, `internal/storage`, `internal/bootstrap`, `internal/session`, and bootstrap tests.
- This change is a bootstrap-ownership foundation, not the full runtime ownership migration.
