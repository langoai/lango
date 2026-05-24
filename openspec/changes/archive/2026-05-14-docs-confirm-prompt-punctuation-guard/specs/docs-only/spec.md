## MODIFIED Requirements

### Requirement: Public CLI prompt examples match current punctuation
Public CLI examples SHALL mirror the current prompt punctuation emitted by the commands they document.

#### Scenario: Shared confirmation prompt punctuation is guarded by tests
- **WHEN** a public doc or README reintroduces a stale shared confirmation example such as `[y/N] y`
- **THEN** the repository test suite SHALL fail
