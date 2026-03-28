## ADDED Requirements

### Requirement: FeatureStatus shared type
The system SHALL provide a `types.FeatureStatus` struct with fields: `Name` (string), `Enabled` (bool), `Healthy` (bool), `Reason` (string), and `Suggestion` (string). This type SHALL be defined in `internal/types/` for cross-layer access.

#### Scenario: Disabled feature with reason
- **WHEN** embedding init skips because no provider is configured
- **THEN** FeatureStatus has `Name: "Embedding & RAG"`, `Enabled: false`, `Healthy: true`, `Reason: "no provider configured"`, `Suggestion: "set embedding.provider or add an OpenAI/Gemini provider"`

#### Scenario: Enabled feature is healthy
- **WHEN** knowledge system initializes successfully
- **THEN** FeatureStatus has `Name: "Knowledge"`, `Enabled: true`, `Healthy: true`, `Reason: ""`, `Suggestion: ""`

### Requirement: StatusCollector aggregation
The system SHALL provide a `StatusCollector` in the app layer that collects `FeatureStatus` from wiring functions. It SHALL expose `All()` to list all statuses and `SilentDisabledCount()` to count features that are disabled with a non-empty reason.

#### Scenario: Silent disabled count
- **WHEN** StatusCollector has 3 features: knowledge (enabled), embedding (disabled, reason="no provider"), graph (disabled, reason="")
- **THEN** `SilentDisabledCount()` returns 1 (only embedding has a reason)

### Requirement: Wiring functions return FeatureStatus
Each context-related wiring function (`initEmbedding`, `initKnowledge`, `initMemory`, `initGraph`, `initLibrarian`) SHALL return a `*types.FeatureStatus` as an additional return value alongside existing components.

#### Scenario: initEmbedding returns status on skip
- **WHEN** `embedding.provider` is empty
- **THEN** `initEmbedding` returns `nil` components AND a non-nil `*types.FeatureStatus` with `Enabled: false` and actionable `Reason`

### Requirement: CLI adapters for FeatureStatus
The CLI layer SHALL provide adapter functions: `FeatureStatusToFeatureInfo` (for status command) and `FeatureStatusToDoctorResult` (for doctor checks). These adapters SHALL live in their respective CLI packages, not in the app layer.

#### Scenario: FeatureStatus to FeatureInfo conversion
- **WHEN** FeatureStatus has `Name: "Knowledge"`, `Enabled: true`, `Reason: ""`
- **THEN** `FeatureStatusToFeatureInfo` returns `FeatureInfo{Name: "Knowledge", Enabled: true, Detail: ""}`

#### Scenario: Disabled FeatureStatus to FeatureInfo with reason
- **WHEN** FeatureStatus has `Name: "Embedding"`, `Enabled: false`, `Reason: "no provider"`
- **THEN** `FeatureStatusToFeatureInfo` returns `FeatureInfo{Name: "Embedding", Enabled: false, Detail: "no provider"}`
