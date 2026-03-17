## MODIFIED Requirements

### Requirement: Collaborator preset includes RPC URL
The collaborator preset SHALL set `payment.network.rpcUrl` to `https://sepolia.base.org` to match the default ChainID 84532 (Base Sepolia). This ensures the preset passes validation when `payment.enabled` is true.

#### Scenario: Collaborator preset passes validation
- **WHEN** `PresetConfig("collaborator")` is called and the result is passed to `PostLoad()`
- **THEN** validation succeeds because both `payment.enabled` and `payment.network.rpcUrl` are set

#### Scenario: Collaborator preset creates valid profile
- **WHEN** `lango config create test --preset collaborator` is run
- **THEN** the profile is created successfully with `payment.network.rpcUrl` set to `https://sepolia.base.org`
