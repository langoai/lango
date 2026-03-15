# smart-account-init-validation Specification

## Purpose
TBD - created by archiving change tool-discovery-audit-bugfix. Update Purpose after archive.
## Requirements
### Requirement: Smart account initialization
The smart account subsystem SHALL validate config fields via SmartAccountConfig.Validate() before proceeding with component creation. initSmartAccount SHALL accept a *lifecycle.Registry parameter for registering lifecycle-managed components. The disabled category description SHALL list all required and recommended config fields.

#### Scenario: Config validation before init
- **WHEN** smartAccount.enabled is true but entryPointAddress is empty
- **THEN** initSmartAccount logs warning "smart account config incomplete" and returns nil without creating components

#### Scenario: Lifecycle registry parameter
- **WHEN** initSmartAccount is called with a non-nil lifecycle registry and sentinel guard is wired
- **THEN** the session guard is registered as a lifecycle component for graceful shutdown

#### Scenario: CLI deps validation
- **WHEN** CLI initializes smart account deps with missing required fields
- **THEN** initSmartAccountDeps returns error wrapping SmartAccountConfig.Validate() result

#### Scenario: Disabled category message
- **WHEN** smart account subsystem is disabled or initialization fails
- **THEN** disabled category description includes required fields (smartAccount.enabled, payment.enabled, entryPointAddress, factoryAddress, bundlerURL) and recommended (economy.enabled)

