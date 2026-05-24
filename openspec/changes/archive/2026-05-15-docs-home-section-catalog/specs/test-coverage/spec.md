## ADDED Requirements

### Requirement: Docs-home section-catalog guard stays executable
Repository-level regressions that let the docs landing page drop links to top-level documentation sections SHALL be enforced by an executable test.

#### Scenario: Every top-level section index remains linked
- **WHEN** the repository still ships top-level `docs/*/index.md` section pages
- **THEN** an executable repository test SHALL fail if `docs/index.md` stops linking any of those section indexes
- **AND** it SHALL therefore catch omissions affecting top-level sections such as `getting-started`, `architecture`, `cli`, `features`, `security`, `gateway`, `payments`, `automation`, `deployment`, or `development`
