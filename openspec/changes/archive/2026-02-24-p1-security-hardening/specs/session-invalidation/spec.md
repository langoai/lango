# Session Explicit Invalidation

## Overview

Extends the TTL-only `SessionStore` with explicit invalidation, auto-invalidation on security events, and CLI management.

## Invalidation Reasons

| Reason | Trigger |
|--------|---------|
| `logout` | User logout |
| `reputation_drop` | Peer trust score falls below threshold |
| `repeated_failures` | Consecutive tool execution failures (default: 5) |
| `manual_revoke` | CLI `lango p2p session revoke` |
| `security_event` | Generic security event |

## SessionStore Enhancements

### New Methods

- `Invalidate(peerDID, reason)` — marks session invalidated, removes from active map, records history, fires callback
- `InvalidateAll(reason)` — invalidates all active sessions
- `InvalidateByCondition(reason, predicate)` — conditional invalidation
- `InvalidationHistory()` — returns invalidation records
- `SetInvalidationCallback(fn)` — registers callback for invalidation events

### Updated Behavior

`Validate()` now returns `false` for sessions with `Invalidated == true`.

## SecurityEventHandler

Automatic session invalidation based on security events:

- **Consecutive failures**: Tracks per-peer failure count. Auto-invalidates at configurable threshold (default 5). Success resets the counter.
- **Reputation drops**: Listens via `reputation.Store.SetOnChangeCallback()`. Invalidates when score falls below `cfg.P2P.MinTrustScore`.

## Protocol Handler Integration

`SecurityEventTracker` interface on `handler.go`:
- `RecordToolSuccess(peerDID)` called after successful tool execution
- `RecordToolFailure(peerDID)` called after failed tool execution

## CLI Commands

- `lango p2p session list [--json]` — show active sessions
- `lango p2p session revoke --peer-did <did>` — revoke specific session
- `lango p2p session revoke-all` — revoke all sessions
