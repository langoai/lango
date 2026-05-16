## MODIFIED Requirements

### Requirement: Chat formatting helpers stay directly regression-tested
The chat TUI formatting helpers that normalize params and transcript annotations SHALL stay directly covered by unit tests, not only by higher-level renderer tests.

#### Scenario: Helper contracts have direct tests
- **WHEN** the repository verifies chat transcript formatting helpers
- **THEN** it SHALL include direct tests for deterministic key sorting, single-line normalization, param-value formatting, and compact request-id generation
