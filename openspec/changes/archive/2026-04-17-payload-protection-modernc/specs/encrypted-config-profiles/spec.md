## ADDED Requirements

### Requirement: Encrypted config profiles stay app-crypto based
Encrypted config profiles MUST continue using application-level crypto and MUST NOT depend on SQLCipher page encryption.

#### Scenario: Profile storage without page encryption
- **WHEN** a config profile is saved or loaded after SQLCipher runtime support is removed
- **THEN** the profile remains encrypted and decryptable through application-managed crypto
- **AND** the operation does not require SQLCipher-specific database features
