## ADDED Requirements

### Requirement: Stdin-pipe acquisition uses shared raw line reader
The non-interactive stdin-pipe passphrase acquisition path SHALL use the shared raw line reader before applying its passphrase-specific trimming and empty-input checks.

#### Scenario: Passphrase stdin path delegates raw line reading
- **WHEN** `ReadStdinPipeFromReader(...)` reads from injected stdin
- **THEN** it SHALL obtain the raw line through the shared lower-level helper
