## ADDED Requirements

### Requirement: Generic section-index parity guard stays executable
Repository-level regressions that let any public docs section index drift away from the dedicated pages in its own directory SHALL be enforced by an executable test.

#### Scenario: Every section index remains complete
- **WHEN** the repository still ships public section indexes under `docs/*/index.md`
- **THEN** an executable repository test SHALL fail if any such section index stops linking one of its sibling `*.md` pages other than `index.md`
- **AND** it SHALL therefore catch omissions across sections such as `architecture`, `cli`, `features`, `security`, `automation`, `payments`, `gateway`, `deployment`, `development`, or `getting-started`
