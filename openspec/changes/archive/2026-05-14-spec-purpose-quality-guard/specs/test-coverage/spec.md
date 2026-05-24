## ADDED Requirements

### Requirement: Main spec hygiene guards stay executable
Repository-level quality regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Main spec purpose placeholders are rejected
- **WHEN** a main OpenSpec spec reintroduces archive-generated placeholder purpose text
- **THEN** an executable repository test SHALL fail
