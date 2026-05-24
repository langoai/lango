## ADDED Requirements

### Requirement: Envelope remains payload-protection root
The master-key envelope MUST remain the root source of key material for broker-managed payload protection after SQLCipher page encryption is removed.

#### Scenario: Envelope-backed payload protection
- **WHEN** the broker needs key material for payload encryption or decryption
- **THEN** it derives the required key material from the envelope-managed master key
- **AND** it does not derive or apply a SQLCipher page key
