## MODIFIED Requirements

### Requirement: Gateway CLI default address coverage stays executable

Repository-level regressions in gateway-backed CLI default address resolution SHALL be enforced by executable tests.

#### Scenario: Gateway IPv6 formatting remains covered
- **WHEN** gateway URL or listen address helpers receive an IPv6 host
- **THEN** executable tests SHALL fail if the formatted address omits required IPv6 brackets

#### Scenario: Doctor server port IPv6 formatting remains covered
- **WHEN** the doctor server port check receives IPv6 gateway bind hosts
- **THEN** executable tests SHALL fail if the check formats the listen address without required IPv6 brackets
