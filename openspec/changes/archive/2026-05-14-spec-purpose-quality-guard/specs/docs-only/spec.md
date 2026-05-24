## MODIFIED Requirements

### Requirement: Main specs use real purpose summaries
Main specs SHALL replace archive-generated placeholder `Purpose` text with concise summaries that describe the actual scope of the spec.

#### Scenario: Archived placeholder purpose text is guarded by tests
- **WHEN** a main spec reintroduces archived placeholder purpose text
- **THEN** the repository test suite SHALL fail
