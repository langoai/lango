## MODIFIED Requirements

### Requirement: Economy configuration struct
The config package SHALL include an EconomyConfig struct with sub-configs for all 5 subsystems. The struct SHALL use mapstructure tags for viper binding.

#### Scenario: Economy config loaded
- **WHEN** configuration is loaded with economy section
- **THEN** EconomyConfig is populated with Budget, Risk, Negotiate, Escrow, and Pricing sub-configs

### Requirement: Config field in main config
The main Config struct SHALL include an Economy field of type EconomyConfig, enabling `economy.enabled`, `economy.budget.*`, etc. configuration paths.

#### Scenario: Economy disabled by default
- **WHEN** no economy config is provided
- **THEN** economy.enabled defaults to false
