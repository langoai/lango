## ADDED Requirements
### Requirement: Contract CLI docs output-contract guards stay executable
Repository-level contract CLI docs regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Stale contract output docs are rejected
- **WHEN** public contract CLI docs or the main contract interaction spec reintroduce boolean `--output` docs or bare `--output` examples without an explicit format
- **THEN** an executable repository test SHALL fail
