## ADDED Requirements

### Requirement: Payment CLI uses storage factories
Payment CLI setup MUST create spending limiters and payment services through storage facade factories instead of generic Ent client access.

#### Scenario: Payment dependencies initialized through facade
- **WHEN** a payment CLI subcommand initializes its dependencies
- **THEN** it obtains the spending limiter and payment service from storage facade factories
