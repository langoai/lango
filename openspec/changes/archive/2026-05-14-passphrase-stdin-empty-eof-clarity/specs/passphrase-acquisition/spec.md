## MODIFIED Requirements

### Requirement: Passphrase acquisition priority chain
The system SHALL acquire a passphrase using the following priority: (1) hardware keyring (Touch ID / TPM), (2) keyfile at `~/.lango/keyfile`, (3) interactive terminal prompt, (4) stdin pipe. The system SHALL return an error if no source is available.

#### Scenario: Empty stdin pipe is treated as empty passphrase input
- **WHEN** the stdin-pipe passphrase path reaches EOF without reading any passphrase bytes
- **THEN** it SHALL return an empty-input passphrase error instead of a raw read failure
