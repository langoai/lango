## ADDED Requirements

### Requirement: Exportability operator docs
The security documentation set SHALL include an exportability document that describes source classes, artifact-level evaluation, decision states, and receipt-style decision records for the first slice.

#### Scenario: Exportability doc available
- **WHEN** a user reads the security documentation
- **THEN** they SHALL find a dedicated exportability document describing the first-slice policy model and its current limits

## MODIFIED Requirements

### Requirement: Security index includes new layers
The `docs/security/index.md` SHALL list OS Keyring, Database Encryption, Cloud KMS/HSM, P2P Session Management, P2P Tool Sandbox, and P2P Auth Hardening in the Security Layers table. It SHALL also link to the exportability operator doc as part of the security documentation set.

#### Scenario: Security layers table updated
- **WHEN** a user reads the security index
- **THEN** they see all 10 security layers including the 6 new ones

#### Scenario: Exportability docs linked from index
- **WHEN** a user reads `docs/security/index.md`
- **THEN** they SHALL find a quick link to the exportability document
