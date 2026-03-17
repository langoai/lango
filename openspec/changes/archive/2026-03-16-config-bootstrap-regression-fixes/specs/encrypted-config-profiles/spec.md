## MODIFIED Requirements

### Requirement: Save normalizes and validates before persisting
The `Store.Save()` method SHALL call `config.PostLoad()` on the config before marshaling and encrypting. This ensures the persisted form is always in canonical form (paths normalized, env vars expanded, validation passed). The mutation of the config is intentional.

#### Scenario: Save normalizes paths before storing
- **WHEN** `Store.Save()` is called with a config containing tilde paths
- **THEN** PostLoad normalizes the paths and the stored config contains absolute paths

#### Scenario: Save rejects invalid config
- **WHEN** `Store.Save()` is called with a config that fails validation (e.g., payment enabled without rpcUrl)
- **THEN** Save returns an error without persisting the config

#### Scenario: Save is safe for already-normalized configs
- **WHEN** `Store.Save()` is called with a config that already passed PostLoad
- **THEN** the double PostLoad call succeeds (idempotent) and the config is stored correctly
