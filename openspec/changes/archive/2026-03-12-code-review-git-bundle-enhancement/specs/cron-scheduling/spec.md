## ADDED Requirements

### Requirement: Scheduler config struct
The `Scheduler` constructor SHALL accept a `SchedulerConfig` struct for optional parameters instead of positional arguments. The constructor signature SHALL be `New(store Store, executor *Executor, cfg SchedulerConfig) *Scheduler`.

#### Scenario: Default config values
- **WHEN** `New()` is called with a zero-value `SchedulerConfig`
- **THEN** defaults SHALL be applied: Timezone="UTC", MaxJobs=5, DefaultTimeout=30m, Logger=zap.NewNop().Sugar()

#### Scenario: Custom config values
- **WHEN** `New()` is called with a populated `SchedulerConfig`
- **THEN** the provided values SHALL override the defaults
