## ADDED Requirements

### Requirement: FTS5 availability in context health
The context health check SHALL report FTS5 availability as a diagnostic detail. When FTS5 is available, the health check SHALL note "FTS5 search active". When FTS5 is unavailable, the health check SHALL note "FTS5 unavailable, using LIKE fallback" as an informational finding (not an error or warning).

#### Scenario: FTS5 available reported in health check
- **WHEN** the context health check runs and FTS5 was successfully probed
- **THEN** the diagnostic output SHALL include "FTS5 search active" as informational detail

#### Scenario: FTS5 unavailable reported in health check
- **WHEN** the context health check runs and FTS5 probe returned false
- **THEN** the diagnostic output SHALL include "FTS5 unavailable, using LIKE fallback" as informational detail
- **AND** this SHALL NOT be reported as a failure or warning (LIKE fallback is a valid operating mode)

#### Scenario: FTS5 status visible in CLI status
- **WHEN** the `lango status` command displays context features
- **THEN** the knowledge feature detail SHALL include whether FTS5 is active or using fallback
