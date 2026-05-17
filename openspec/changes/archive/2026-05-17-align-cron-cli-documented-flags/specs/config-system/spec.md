## MODIFIED Requirements

### Requirement: Cron configuration
The config system SHALL support a `cron` section with fields: enabled (bool),
timezone (string), maxConcurrentJobs (int), defaultSessionMode (string),
historyRetention (duration string), defaultJobTimeout (duration string),
defaultDeliverTo ([]string).

#### Scenario: Default cron config
- **WHEN** no cron config is specified
- **THEN** defaults SHALL be: enabled=false, timezone="UTC",
  maxConcurrentJobs=5, defaultSessionMode="isolated",
  historyRetention="720h", defaultJobTimeout="30m", defaultDeliverTo=nil
