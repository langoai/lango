## MODIFIED Requirements

### Requirement: ProviderConfig type strengthening
The `ProviderConfig.Type` field SHALL use `types.ProviderType` instead of raw `string`.

#### Scenario: Config deserialization with typed provider
- **WHEN** config is loaded via mapstructure/viper
- **THEN** `ProviderConfig.Type` SHALL deserialize correctly as `types.ProviderType`

#### Scenario: Provider validation
- **WHEN** a `ProviderConfig` is created with an unknown provider type
- **THEN** `config.Type.Valid()` SHALL return `false`
