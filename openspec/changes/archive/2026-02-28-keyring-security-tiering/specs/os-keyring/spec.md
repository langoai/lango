## MODIFIED Requirements

### Requirement: Status struct describes keyring availability
The `Status` struct SHALL include a `SecurityTier` field indicating the detected hardware security tier alongside existing `Available`, `Backend`, and `Error` fields.

#### Scenario: IsAvailable reports security tier
- **WHEN** `IsAvailable()` is called on a system with biometric hardware
- **THEN** the returned `Status` SHALL have `SecurityTier` set to `TierBiometric`

#### Scenario: IsAvailable on system without secure hardware
- **WHEN** `IsAvailable()` is called on a system without biometric or TPM
- **THEN** the returned `Status` SHALL have `SecurityTier` set to `TierNone`
