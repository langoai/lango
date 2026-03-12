## MODIFIED Requirements

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

#### Scenario: Separate context for git state collection
- **WHEN** a health ping succeeds and a GitStateProvider is configured
- **THEN** the monitor SHALL create a separate timeout context for git state collection, independent of the ping context, so that git state calls are not starved by ping latency

### Requirement: Health monitor start/stop lifecycle
The HealthMonitor SHALL start and stop cleanly with a goroutine-safe lifecycle.

#### Scenario: Start launches check loop
- **WHEN** `Start()` is called
- **THEN** the periodic health check goroutine SHALL begin and event subscriptions SHALL be active

#### Scenario: Stop halts check loop
- **WHEN** `Stop()` is called
- **THEN** the health check goroutine SHALL exit cleanly and the WaitGroup SHALL complete

#### Scenario: Restart does not duplicate subscriptions
- **WHEN** `Start()` is called multiple times (e.g., after a restart)
- **THEN** event bus subscriptions SHALL be registered exactly once via `sync.Once`, preventing duplicate handler accumulation
