## MODIFIED Requirements

### Requirement: Existing Test Enhancements
Existing test files SHALL be expanded with additional scenarios.

#### Scenario: Existing package tests gain targeted enhancements
- **WHEN** the enhanced regression suite runs
- **THEN** it SHALL cover session-store CRUD and TTL behavior, anthropic model listing, openai unavailable-server handling, app startup failure modes, and doctor non-conflict listen failures
