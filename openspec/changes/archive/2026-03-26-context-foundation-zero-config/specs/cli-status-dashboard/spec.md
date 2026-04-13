## MODIFIED Requirements

### Requirement: Status command shows context profile
`lango status` SHALL display the active context profile name alongside existing feature information. If no profile is set, the profile field SHALL show "none" or be omitted.

#### Scenario: Profile shown in status output
- **WHEN** `contextProfile: balanced` is set
- **THEN** `lango status` output includes "Profile: balanced" in the dashboard header or feature section

### Requirement: Feature detail includes reason
The `Detail` field of context-related features in `collectFeatures()` SHALL reflect the `FeatureStatus.Reason` when available, providing users actionable context about why a feature is disabled.

#### Scenario: Embedding detail shows reason
- **WHEN** embedding is disabled because no provider is configured
- **THEN** status output for "Embedding & RAG" shows `Detail: "no provider configured"` instead of empty string

#### Scenario: Enabled feature detail unchanged
- **WHEN** knowledge is enabled and healthy
- **THEN** status output for "Knowledge" shows existing detail behavior (empty or provider info)
