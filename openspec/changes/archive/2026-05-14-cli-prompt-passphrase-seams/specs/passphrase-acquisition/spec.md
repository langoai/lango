## MODIFIED Requirements

### Requirement: Passphrase acquisition priority chain
The system SHALL acquire a passphrase using the following priority: (1) hardware keyring (Touch ID / TPM), (2) keyfile at `~/.lango/keyfile`, (3) interactive terminal prompt, (4) stdin pipe. The system SHALL return an error if no source is available.

#### Scenario: Hidden passphrase prompt supports deterministic test seams
- **WHEN** the hidden-input passphrase prompt is exercised in tests
- **THEN** the prompt output writer, stdin file descriptor, and password reader SHALL be injectable without changing the runtime prompt behavior
