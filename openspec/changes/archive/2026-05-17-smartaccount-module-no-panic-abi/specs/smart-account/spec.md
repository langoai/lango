## MODIFIED Requirements

### Requirement: Module registry
The system SHALL provide a module registry in `internal/smartaccount/module/`:
- `Registry`: Register/List/Get module descriptors
- `ABIEncoder`: Encode installModule/uninstallModule calldata (ERC-7579)
- Pre-registered: LangoSessionValidator, LangoSpendingHook, LangoEscrowExecutor

The ABI encoder SHALL return errors for internal ABI argument initialization failures instead of panicking.

#### Scenario: Pre-registered modules
- **WHEN** the module registry is initialized with module addresses
- **THEN** LangoSessionValidator, LangoSpendingHook, and LangoEscrowExecutor SHALL be pre-registered

#### Scenario: Module ABI encoder avoids panic paths
- **WHEN** installModule or uninstallModule calldata is encoded
- **THEN** the encoder SHALL return either encoded calldata or an error
- **AND** it SHALL NOT panic due to deterministic ABI argument setup
