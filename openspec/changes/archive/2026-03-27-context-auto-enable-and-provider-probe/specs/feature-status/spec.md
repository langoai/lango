## ADDED Requirements

### Requirement: FeatureStatus AutoEnabled field
`FeatureStatus` SHALL include `AutoEnabled bool` field indicating the feature was auto-enabled rather than explicitly configured.

#### Scenario: AutoEnabled in JSON
- **WHEN** a FeatureStatus with `AutoEnabled: true` is serialized to JSON
- **THEN** the output SHALL include `"autoEnabled": true`
