## Why

The storage broker client still has transport-layer correctness bugs: concurrent request writes are unsynchronized, large responses can overflow the default scanner token limit, and `Close()` never actually sends the graceful shutdown RPC. These issues can corrupt the stdio protocol or bypass intended shutdown cleanup even though higher-level tests pass.

## What Changes

- Serialize broker request writes with a dedicated write mutex.
- Replace scanner-based response reading with a decoder path that supports large payload responses.
- Fix `Close()` so it attempts a real shutdown RPC before marking the client closed.
- Add transport-focused regression tests for concurrent writes, large responses, and shutdown behavior.

## Capabilities

### Modified Capabilities
- `brokered-storage`: persistent stdio JSON transport is made concurrency-safe, large-payload-safe, and graceful-shutdown-correct.

## Impact

- Affected code: `internal/storagebroker/client.go` and transport-focused tests.
- No user-facing behavior change intended beyond eliminating protocol corruption and shutdown edge cases.
