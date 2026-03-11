# Proposal: P2P Team Management Enhancements

## Why

The P2P team coordination system lacked persistent storage, health monitoring, membership management, and graceful shutdown. Teams existed only in memory and could not survive restarts, unhealthy members went undetected, and there was no orderly way to shut down a team with escrow cleanup.

## What Changes

- **BoltDB Persistent Team Store**: New `TeamStore` interface and `BoltStore` implementation for persisting team state across restarts via BoltDB. The `Coordinator` now accepts a `TeamStore` and auto-persists on formation, task completion, member kicks, and disbanding.
- **Coordinator Membership Management**: New `KickMember()` method for removing members with reason and event publishing. New `TeamsForMember()` query to find all active teams containing a given DID.
- **Graceful Team Shutdown**: New `GracefulShutdown()` method that transitions teams through `StatusShuttingDown` (blocking new tasks), calculates proportional settlement, publishes `TeamGracefulShutdownEvent`, and then disbands.
- **Health Monitor**: New `HealthMonitor` component that periodically pings team members via `health_ping` invocations, tracks consecutive misses, and publishes `TeamMemberUnhealthyEvent` when threshold is exceeded. Subscribes to task completion events to reset counters. Implements `lifecycle.Component` for clean start/stop.
- **Enhanced Team Model**: Added `Budget`, `Spent`, `StatusShuttingDown`, `StatusCompleted` fields. Added `Members()`, `MemberCount()`, `ActiveMembers()`, `AddSpend()` methods. Added JSON marshal/unmarshal with members slice serialization. Added `RoleReviewer` constant.
- **New Event Types**: Added `TeamHealthCheckEvent`, `TeamMemberUnhealthyEvent`, `TeamBudgetWarningEvent`, `TeamGracefulShutdownEvent`, `TeamLeaderChangedEvent`.
- **Configuration**: Added `TeamConfig` with `HealthCheckInterval`, `MaxMissedHeartbeats`, `MinReputationScore` to `P2PConfig`.

## Capabilities

### New Capabilities
- `team-health-monitoring`: Periodic health checking of team members with auto-detection of unhealthy members

### Modified Capabilities
- `p2p-team-coordination`: BoltDB persistent store, coordinator membership management, graceful shutdown, enhanced team model with budget tracking

## Impact

- **Core packages modified**: `internal/p2p/team/`, `internal/eventbus/`, `internal/config/`
- **New files**: `bolt_store.go`, `coordinator_membership.go`, `coordinator_shutdown.go`, `health_monitor.go`, `integration_test.go`, plus test files
- **Dependencies**: `go.etcd.io/bbolt` (already in use for other stores)
- **No breaking changes**: All additions are backward-compatible
