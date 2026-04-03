## MODIFIED Requirements

### Requirement: Session eviction policy
When `MaxSessions <= 0`, the `evictOldestSession` method MUST skip eviction entirely (unlimited sessions). When `MaxSessions > 0`, eviction MUST occur when the session count reaches the configured limit.

#### Scenario: Unlimited sessions
- **WHEN** `MaxSessions` is 0
- **THEN** `evictOldestSession` SHALL return immediately without removing any session

#### Scenario: Eviction at capacity
- **WHEN** `MaxSessions` is 100 and 100 sessions exist
- **THEN** the oldest session (by `LastUpdated`) SHALL be evicted before adding a new one
