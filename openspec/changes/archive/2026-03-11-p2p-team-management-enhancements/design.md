# Design: P2P Team Management Enhancements

## Context

The P2P team coordinator managed teams in memory only. Teams were lost on restart, there was no way to detect unresponsive members, no mechanism for orderly shutdown with escrow settlement, and no membership management beyond initial formation. The team model also lacked budget tracking fields needed for the team-escrow bridge.

## Goals / Non-Goals

**Goals:**
- Persist team state to BoltDB so teams survive process restarts
- Detect unhealthy team members via periodic health pings
- Support orderly team shutdown that blocks new tasks and publishes settlement events
- Enable runtime membership management (kick members, query membership)
- Extend the team model with budget, spend, and lifecycle state fields

**Non-Goals:**
- Auto-removal of unhealthy members (event is published; the bridge layer decides action)
- Leader election or automatic leader failover
- Cross-node team state replication (single-node BoltDB is sufficient)
- CLI commands for health monitoring (health data is event-driven, consumed by bridges)

## Decisions

### BoltDB for team persistence (not Ent/SQLite)
Teams are ephemeral, task-scoped entities with simple key-value access patterns. BoltDB is already used for other P2P stores (reputation, workspace) and avoids the overhead of SQL migrations. The `TeamStore` interface allows swapping to a different backend later.

**Alternative considered:** Ent/SQLite would provide query capabilities but adds migration complexity for data that is primarily accessed by ID. Teams are short-lived (minutes to hours), not long-lived records.

### Interface-based store design
The `TeamStore` interface (Save/Load/LoadAll/Delete) decouples the coordinator from the storage backend. This enables in-memory stores for testing and alternative backends without modifying coordination logic.

### HealthMonitor as lifecycle.Component
The health monitor runs a periodic goroutine and needs clean startup/shutdown. Implementing `lifecycle.Component` (Name/Start/Stop) integrates it with the existing lifecycle registry for ordered startup and graceful shutdown.

**Alternative considered:** Using a cron job for health checks. Rejected because the health monitor needs persistent counter state and event subscriptions that don't map cleanly to the cron model.

### Event-driven counter reset
Instead of querying the coordinator for task completion status, the health monitor subscribes to `TeamTaskCompletedEvent` via the EventBus. This avoids coupling and reuses the existing event pattern. Similarly, `TeamDisbandedEvent` triggers cleanup of tracking maps to prevent memory leaks.

### StatusShuttingDown as explicit state
Adding a `StatusShuttingDown` team state (between Active and Disbanded) enables `DelegateTask` to reject new work during graceful shutdown. This is simpler than a separate shutdown flag and integrates cleanly with existing state checks.

### Members serialized as slice in JSON
The internal `map[string]*Member` is serialized as a JSON array via custom `MarshalJSON`/`UnmarshalJSON`. This produces cleaner JSON output and is straightforward to reconstruct into the map on deserialization.

## Risks / Trade-offs

- **[Risk] BoltDB single-writer bottleneck** -> Team operations are infrequent (formation/disbanding) so write contention is negligible. Read-heavy operations (GetTeam, ListTeams) use the in-memory map, not the store.
- **[Risk] Health ping false positives under network partitions** -> The maxMissed threshold (default 3) with 30s intervals gives 90 seconds before unhealthy detection, providing tolerance for transient network issues.
- **[Risk] Memory growth from tracking maps** -> The `TeamDisbandedEvent` subscription in HealthMonitor cleans up miss counts and lastSeen maps when teams end. Active teams are bounded by operational use.
- **[Trade-off] No auto-kick on unhealthy** -> Publishing `TeamMemberUnhealthyEvent` rather than auto-removing gives the bridge layer control over the response (e.g., retry, degrade, or kick). This is more flexible but requires external handling.

## Files

| File | Type | Purpose |
|------|------|---------|
| `internal/p2p/team/bolt_store.go` | New | BoltDB-backed TeamStore implementation |
| `internal/p2p/team/bolt_store_test.go` | New | Comprehensive store tests |
| `internal/p2p/team/team.go` | Modified | Budget, spend, status, member methods, JSON serialization |
| `internal/p2p/team/coordinator.go` | Modified | Store integration, GetTeam, ListTeams, LoadPersistedTeams, event publishing |
| `internal/p2p/team/coordinator_membership.go` | New | KickMember, TeamsForMember |
| `internal/p2p/team/coordinator_shutdown.go` | New | GracefulShutdown with StatusShuttingDown |
| `internal/p2p/team/health_monitor.go` | New | Periodic health check with miss tracking |
| `internal/p2p/team/health_monitor_test.go` | New | Health check logic tests |
| `internal/p2p/team/integration_test.go` | New | End-to-end team lifecycle tests |
| `internal/eventbus/team_events.go` | Modified | New event types for health, shutdown, budget, leader |
| `internal/config/types_p2p.go` | Modified | TeamConfig struct |
