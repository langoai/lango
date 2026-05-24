## MODIFIED Requirements

### Requirement: Payment CLI uses storage factories
Payment CLI setup MUST create spending limiters and payment services through storage-facing payment capabilities instead of direct Ent client access.

#### Scenario: Payment dependencies initialized through facade
- **WHEN** a payment CLI subcommand initializes its dependencies
- **THEN** it obtains the spending limiter and payment transaction persistence from storage-facing capabilities
- **AND** payment service construction stays in the payment/CLI layer
- **AND** it does not extract or consume a raw `*ent.Client` from production storage paths
