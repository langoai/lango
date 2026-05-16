## ADDED Requirements

### Requirement: README advanced setup paths stay truth-aligned
README guidance for advanced configuration SHALL point only to setup paths that exist in the current product surfaces.

#### Scenario: README avoids false onboard submenu navigation
- **WHEN** a user reads README guidance for prompts, embedding, graph, multi-agent, A2A, security mode, or OIDC auth
- **THEN** the document SHALL point to real setup paths such as `lango settings` or `lango config import/export`
- **AND** SHALL NOT describe nonexistent advanced onboard submenu navigation

### Requirement: Advanced feature docs avoid false onboard submenu flows
Downstream feature documentation SHALL describe advanced feature setup using the actual onboarding and settings surfaces.

#### Scenario: Feature docs keep advanced setup paths synchronized
- **WHEN** a user reads advanced feature docs for embedding/RAG, A2A, or the knowledge graph
- **THEN** those docs SHALL describe `lango settings` and/or config import/export as the interactive setup path
- **AND** SHALL clarify that the five-step onboard wizard is only the initial bootstrap flow
