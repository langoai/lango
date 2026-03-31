## MODIFIED Requirements

### Requirement: Ontology surface tools
The system SHALL provide 13 static agent-facing tools with `ontology_` prefix plus dynamic `ontology_action_{name}` tools generated from the ActionRegistry and an `ontology_list_actions` tool. `BuildTools` SHALL accept `(svc OntologyService, reg *ActionRegistry)` — when reg is nil, only static tools are returned. Read-only tools SHALL use SafetyLevelSafe; mutation tools SHALL use SafetyLevelModerate.

#### Scenario: BuildTools with nil registry
- **WHEN** `BuildTools(svc, nil)` is called
- **THEN** it SHALL return exactly 13 static ontology tools

#### Scenario: BuildTools with registry containing 2 actions
- **WHEN** `BuildTools(svc, reg)` is called with a registry containing 2 registered actions
- **THEN** it SHALL return 13 static + 1 list_actions + 2 dynamic action tools = 16 tools

#### Scenario: Dynamic action tool name format
- **WHEN** an action named "link_entities" is registered
- **THEN** a tool named `ontology_action_link_entities` SHALL be generated

### Requirement: Ontologist identity prompt
The system SHALL provide a `prompts/agents/ontologist/IDENTITY.md` file defining the ontologist agent's role, capabilities, and tool usage guidelines. The identity prompt SHALL include a note that ontology operations may be restricted by ACL permissions based on the calling agent's role. The identity prompt SHALL list `ontology_list_actions` and `ontology_action_*` dynamic tools.

#### Scenario: Identity prompt lists action tools
- **WHEN** ontologist agent identity is loaded
- **THEN** the prompt SHALL contain `ontology_list_actions` and `ontology_action_*` tool descriptions
