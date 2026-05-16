## ADDED Requirements
### Requirement: Contract CLI docs stay aligned with the explicit output-format contract
Public contract CLI docs and the main contract interaction spec SHALL keep the current `--output table|json` contract instead of drifting back to a boolean output toggle.

#### Scenario: Stale contract boolean-output docs are rejected
- **WHEN** public contract CLI docs or the main contract interaction spec reintroduce a boolean `--output` flag table entry or a bare `--output` example without an explicit format
- **THEN** an executable repository test SHALL fail
