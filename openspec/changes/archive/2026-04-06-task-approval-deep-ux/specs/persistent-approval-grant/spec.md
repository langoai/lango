## MODIFIED Requirements

### Requirement: Grant store enumeration
The GrantStore SHALL provide a List() method returning all active (non-expired) grants as []GrantInfo sorted by session key then tool name. Expired grants SHALL be excluded.

#### Scenario: List active grants
- **WHEN** 3 grants exist and none are expired
- **THEN** List() returns 3 GrantInfo sorted by session then tool

#### Scenario: Expired grants excluded
- **WHEN** a grant has expired based on the TTL
- **THEN** List() excludes it from the result

#### Scenario: Revoked grants excluded
- **WHEN** a grant has been revoked
- **THEN** List() excludes it from the result
