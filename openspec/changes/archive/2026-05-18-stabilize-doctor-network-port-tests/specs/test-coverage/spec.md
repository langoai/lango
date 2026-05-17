## MODIFIED Requirements

### Requirement: Existing Test Enhancements
Existing test files SHALL be expanded with additional scenarios and SHALL avoid fixed-port assumptions where the OS can allocate an available loopback port.

#### Scenario: Existing package tests gain targeted enhancements
- **WHEN** the enhanced regression suite runs
- **THEN** it SHALL cover session-store CRUD and TTL behavior, anthropic model listing, openai unavailable-server handling, app startup failure modes, doctor non-conflict listen failures, and doctor port-available checks without fixed-port assumptions
