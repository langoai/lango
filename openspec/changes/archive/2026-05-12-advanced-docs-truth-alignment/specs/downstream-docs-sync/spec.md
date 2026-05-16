## ADDED Requirements

### Requirement: Advanced configuration docs avoid false onboarding and file-path guidance
Public documentation for advanced configuration SHALL point to the actual interactive and programmatic configuration paths in the current product.

#### Scenario: Advanced docs use settings or import-export paths
- **WHEN** a user reads advanced configuration guidance in README or feature docs
- **THEN** that guidance SHALL point to `lango settings` and/or `lango config import/export`
- **AND** SHALL NOT claim that advanced feature setup happens through nonexistent `lango onboard` submenu paths
- **AND** SHALL NOT describe a canonical plaintext `~/.lango/config.yaml` configuration file
