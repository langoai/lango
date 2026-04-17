## ADDED Requirements

### Requirement: Status reflects brokered payload protection
The security status surface MUST report brokered payload protection state rather than SQLCipher page-encryption state once the new protection model is active.

#### Scenario: Payload protection status reporting
- **WHEN** the user runs the security status command after payload protection is enabled
- **THEN** the output reports broker/storage/payload-protection state
- **AND** it does not imply that SQLCipher page encryption is active
