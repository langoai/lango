## MODIFIED Requirements

### Requirement: Configuration validation uses exported constants
The config `Validate()` function SHALL reference exported package-level validation maps (`ValidLogLevels`, `ValidLogFormats`, `ValidSignerProviders`, `ValidWalletProviders`, `ValidZKPSchemes`, `ValidContainerRuntimes`, `ValidMCPTransports`) instead of inline map literals.

#### Scenario: Validation map reuse
- **WHEN** `config.Validate()` checks the log level value
- **THEN** it uses `config.ValidLogLevels` map defined in `constants.go`

#### Scenario: External access to valid values
- **WHEN** another package needs to validate a config value (e.g., CLI flag validation)
- **THEN** it can import and use `config.ValidLogLevels` directly
