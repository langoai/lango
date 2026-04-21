## ADDED Requirements

### Requirement: Security index reflects the public deep-dive slice
The `docs/security/index.md` SHALL provide quick links to the public security deep-dive docs surfaced by the MkDocs IA recovery slice, including Approval CLI and Envelope Migration.

#### Scenario: Approval CLI quick link is present
- **WHEN** a user reads `docs/security/index.md`
- **THEN** they find a quick link to `approval-cli.md`

#### Scenario: Envelope Migration quick link is present
- **WHEN** a user reads `docs/security/index.md`
- **THEN** they find a quick link to `envelope-migration.md`
