## MODIFIED Requirements

### Requirement: Ontology surface tools
The system SHALL provide 13 static agent-facing tools with `ontology_` prefix, plus dynamic `ontology_action_{name}` tools generated from the ActionRegistry, an `ontology_list_actions` tool, and 4 governance tools: `ontology_promote_type`, `ontology_promote_predicate`, `ontology_schema_health`, `ontology_type_usage`. `BuildTools` SHALL accept `(svc OntologyService, reg *ActionRegistry)`. Read-only tools SHALL use SafetyLevelSafe; mutation tools SHALL use SafetyLevelModerate.

#### Scenario: Governance tools present
- **WHEN** `BuildTools` is called
- **THEN** the result SHALL include `ontology_promote_type`, `ontology_promote_predicate`, `ontology_schema_health`, `ontology_type_usage`

### Requirement: Ontologist identity prompt
The system SHALL provide a `prompts/agents/ontologist/IDENTITY.md` file defining the ontologist agent's role, capabilities, and tool usage guidelines. The identity prompt SHALL include a note that ontology operations may be restricted by ACL permissions based on the calling agent's role. The identity prompt SHALL list `ontology_list_actions`, `ontology_action_*` dynamic tools, and governance tools (`ontology_promote_type`, `ontology_promote_predicate`, `ontology_schema_health`, `ontology_type_usage`).

#### Scenario: Identity prompt lists governance tools
- **WHEN** ontologist agent identity is loaded
- **THEN** the prompt SHALL contain governance tool descriptions
