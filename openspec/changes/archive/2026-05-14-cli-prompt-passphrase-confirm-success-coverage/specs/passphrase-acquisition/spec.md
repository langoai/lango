## MODIFIED Requirements

### Requirement: Passphrase acquisition priority chain
The system SHALL acquire a passphrase using the following priority: (1) hardware keyring (Touch ID / TPM), (2) keyfile at `~/.lango/keyfile`, (3) interactive terminal prompt, (4) stdin pipe. The system SHALL return an error if no source is available.

#### Scenario: New passphrase creation success path is deterministic under test
- **WHEN** the interactive new-passphrase confirmation flow is exercised in tests
- **THEN** the hidden-input reads, prompt output sequence, and returned passphrase SHALL be assertable through the existing prompt seams
