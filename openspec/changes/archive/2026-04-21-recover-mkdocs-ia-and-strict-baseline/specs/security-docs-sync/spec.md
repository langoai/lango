## ADDED Requirements

### Requirement: Security index reflects the public deep-dive slice
The `docs/security/index.md` SHALL surface the newly public deep-dive docs exposed by the MkDocs IA recovery slice, including Approval CLI and Envelope Migration.

#### Scenario: Approval CLI is surfaced from the security index
- **WHEN** a user reads `docs/security/index.md`
- **THEN** they find a link or quick reference to `approval-cli.md`

#### Scenario: Envelope Migration is surfaced from the security index
- **WHEN** a user reads `docs/security/index.md`
- **THEN** they find a link or quick reference to `envelope-migration.md`
