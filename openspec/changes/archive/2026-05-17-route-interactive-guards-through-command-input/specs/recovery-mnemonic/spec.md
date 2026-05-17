## MODIFIED Requirements

### Requirement: Recovery restore requires interactive terminal
`lango security recovery restore` SHALL require an interactive command input
stream before prompting for the mnemonic or replacement passphrase.

#### Scenario: Non-interactive restore fails before prompting
- **WHEN** `lango security recovery restore` is run with a non-interactive
  command input stream
- **THEN** it SHALL return an error requiring an interactive terminal
- **AND** it SHALL NOT prompt for the mnemonic or replacement passphrase

#### Scenario: Recovery restore guard uses command input
- **WHEN** `lango security recovery restore` reaches its interactive guard
- **THEN** the guard SHALL validate the command input stream instead of reading
  process-global stdin directly

## ADDED Requirements

### Requirement: Recovery setup guard uses command input stream
`lango security recovery setup` SHALL validate the Cobra command input stream
before prompting for the current passphrase or mnemonic confirmation.

#### Scenario: Recovery setup guard uses command input
- **WHEN** `lango security recovery setup` reaches its interactive guard
- **THEN** the guard SHALL validate the command input stream instead of reading
  process-global stdin directly
