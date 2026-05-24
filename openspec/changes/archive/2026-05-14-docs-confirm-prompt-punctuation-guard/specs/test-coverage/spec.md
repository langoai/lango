## ADDED Requirements

### Requirement: Public doc shared-confirm punctuation guards stay executable
Repository-level docs-quality regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Public doc shared-confirm punctuation regressions are rejected
- **WHEN** a public doc or README reintroduces stale shared confirmation examples without the colon separator
- **THEN** an executable repository test SHALL fail
