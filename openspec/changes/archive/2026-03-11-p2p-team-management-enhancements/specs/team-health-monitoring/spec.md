## ADDED Requirements

### Requirement: Health Monitor component
The `team` package SHALL provide a `HealthMonitor` that periodically pings team members and detects unhealthy members based on consecutive missed pings.

#### Scenario: Health monitor creation
- **WHEN** `NewHealthMonitor` is called with a config
- **THEN** it SHALL use the configured interval (default 30s) and maxMissed threshold (default 3)

#### Scenario: Implements lifecycle.Component
- **WHEN** the HealthMonitor is registered in the lifecycle registry
- **THEN** it SHALL implement `Name()`, `Start(ctx, wg)`, and `Stop(ctx)` methods

### Requirement: Periodic health checks
The HealthMonitor SHALL ping all non-leader active members of active teams at the configured interval.

#### Scenario: Ping active workers
- **WHEN** the health check interval elapses
- **THEN** the monitor SHALL send `health_ping` to all active non-leader members of all active teams concurrently

#### Scenario: Successful ping resets counter
- **WHEN** a member responds to a health ping successfully
- **THEN** the miss counter for that member SHALL be reset to zero and the lastSeen timestamp SHALL be updated

#### Scenario: Failed ping increments counter
- **WHEN** a member fails to respond to a health ping
- **THEN** the consecutive miss counter SHALL be incremented

#### Scenario: Unhealthy detection
- **WHEN** a member's consecutive miss count reaches or exceeds the maxMissed threshold
- **THEN** a `TeamMemberUnhealthyEvent` SHALL be published with the member's DID, name, miss count, and lastSeen timestamp

### Requirement: Aggregate health events
The HealthMonitor SHALL publish a `TeamHealthCheckEvent` after each team-level sweep.

#### Scenario: Health check event published
- **WHEN** a team health sweep completes
- **THEN** a `TeamHealthCheckEvent` SHALL be published with the count of healthy members and total members checked

### Requirement: Counter management via events
The HealthMonitor SHALL subscribe to EventBus events to manage its internal counters.

#### Scenario: Task completion resets counters
- **WHEN** a `TeamTaskCompletedEvent` is received
- **THEN** all miss counters for that team SHALL be reset and lastSeen timestamps SHALL be updated

#### Scenario: Team disbanded cleans up
- **WHEN** a `TeamDisbandedEvent` is received
- **THEN** all tracking data (miss counts and lastSeen maps) for that team SHALL be removed to prevent memory leaks

### Requirement: Health monitor start/stop lifecycle
The HealthMonitor SHALL start and stop cleanly with a goroutine-safe lifecycle.

#### Scenario: Start launches check loop
- **WHEN** `Start()` is called
- **THEN** the periodic health check goroutine SHALL begin and event subscriptions SHALL be active

#### Scenario: Stop halts check loop
- **WHEN** `Stop()` is called
- **THEN** the health check goroutine SHALL exit cleanly and the WaitGroup SHALL complete
