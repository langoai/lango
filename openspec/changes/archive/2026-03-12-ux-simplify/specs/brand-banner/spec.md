## MODIFIED Requirements

### Requirement: Startup feature summary
The banner package SHALL provide a `StartupSummary(features []FeatureLine) string` function that renders a list of activated/deactivated features using checkmark/cross indicators. Each FeatureLine has Name (string), Enabled (bool), and Detail (string) fields.

#### Scenario: Mixed features
- **WHEN** StartupSummary is called with Gateway(enabled, "http://localhost:18789") and P2P(disabled)
- **THEN** output contains a pass-formatted Gateway line with detail and a fail-formatted P2P line

#### Scenario: Empty features
- **WHEN** StartupSummary is called with empty slice
- **THEN** output is empty string

### Requirement: Version accessor
The banner package SHALL expose `GetVersion() string` returning the package-level version string.

#### Scenario: Default version
- **WHEN** GetVersion() is called without SetVersionInfo
- **THEN** returns "dev"
