## ADDED Requirements
### Requirement: Architecture docs stay aligned with current code paths
Architecture docs SHALL not point to deleted implementation files when the current code moved to a new path.

#### Scenario: Stale librarian proactive buffer path is rejected
- **WHEN** the architecture data-flow page documents the proactive librarian buffer
- **THEN** it SHALL reference the current implementation path instead of a deleted `internal/librarian/buffer.go`
