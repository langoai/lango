## ADDED Requirements
### Requirement: Public config CLI examples remain profile-aware and flag-accurate
Public docs and README examples for config import/export/get SHALL remain aligned with the real CLI contract.

#### Scenario: Stale config CLI examples are rejected
- **WHEN** a public doc or README reintroduces `lango config get ... --format json`, `lango config export` without a profile argument, or `lango config import` without an explicit `--profile`
- **THEN** an executable repository test SHALL fail
