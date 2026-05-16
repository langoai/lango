## ADDED Requirements

### Requirement: Main specs use real purpose summaries
Main specs SHALL replace archive-generated placeholder `Purpose` text with concise summaries that describe the actual scope of the spec.

#### Scenario: Archived placeholder purpose is removed
- **WHEN** a main spec still contains placeholder text left by an archived change
- **THEN** the spec SHALL be updated to use a real purpose summary
