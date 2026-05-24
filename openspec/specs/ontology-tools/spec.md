## Purpose

Capability spec for ontology-tools. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: Ontology surface tools
The system SHALL provide 13 static agent-facing tools with `ontology_` prefix, plus dynamic `ontology_action_{name}` tools generated from the ActionRegistry, an `ontology_list_actions` tool, and 4 governance tools: `ontology_promote_type`, `ontology_promote_predicate`, `ontology_schema_health`, `ontology_type_usage`. `BuildTools` SHALL accept `(svc OntologyService, reg *ActionRegistry)`. Read-only tools SHALL use SafetyLevelSafe; mutation tools SHALL use SafetyLevelModerate.

#### Scenario: Ontology tool builder exposes static, dynamic, and governance tools
- **WHEN** ontology tools are built from the service and action registry
- **THEN** the resulting toolset MUST include the documented static tools, dynamic `ontology_action_*` tools, `ontology_list_actions`, and the governance tools

### Requirement: Ontologist agent routing
The system SHALL define an `ontologist` AgentSpec with Prefixes `["ontology_"]`. All `ontology_` prefixed tools SHALL be routed to the ontologist agent via PartitionTools.

#### Scenario: Ontology-prefixed tools route to the ontologist
- **WHEN** tool partitioning evaluates an `ontology_`-prefixed tool
- **THEN** the tool MUST route to the `ontologist` agent

### Requirement: Ontology tools registration
When `ontology.enabled` is true and OntologyService is initialized, the system SHALL register ontology tools via CatalogEntry in the intelligence module.

#### Scenario: Enabled ontology service registers ontology tools
- **WHEN** ontology is enabled and the ontology service initializes successfully
- **THEN** the intelligence module MUST register ontology tools through catalog entries

### Requirement: JSON import tool
`ontology_import_json` SHALL accept a `data` parameter (JSON string) containing entities with id, type, properties, and optional relations. Each entity SHALL be validated via SetEntityProperty (type+property schema) and relations via AssertFact (predicate validation).

#### Scenario: JSON import validates entities and relations
- **WHEN** `ontology_import_json` imports JSON entity data
- **THEN** entity properties MUST be validated through `SetEntityProperty`
- **AND** relations MUST be validated through `AssertFact`

### Requirement: CSV import tool
`ontology_import_csv` SHALL accept `data` (CSV string) and `type` (ObjectType name). The first row SHALL be treated as property name headers. Each subsequent row creates an entity with the given type and column values as properties.

#### Scenario: CSV import maps columns to typed properties
- **WHEN** `ontology_import_csv` receives CSV data and an object type
- **THEN** the first row MUST be interpreted as property headers
- **AND** subsequent rows MUST create entities of the given type with those property values

### Requirement: MCP result import tool
`ontology_from_mcp` SHALL accept `tool_name`, `result_json` (JSON string), `entity_type`, and `predicate`. The handler SHALL decode the JSON, create entity properties, and assert a fact linking the entity to the tool via the specified predicate. Explicit mapping only — no automatic type inference.

#### Scenario: MCP import links imported entities to the source tool
- **WHEN** `ontology_from_mcp` imports JSON output from an MCP tool
- **THEN** it MUST decode the JSON into entity properties
- **AND** assert a fact linking the entity to the named tool using the supplied predicate

### Requirement: Ontologist identity prompt
The system SHALL provide a `prompts/agents/ontologist/IDENTITY.md` file defining the ontologist agent's role, capabilities, and tool usage guidelines. The identity prompt SHALL include a note that ontology operations may be restricted by ACL permissions based on the calling agent's role. The identity prompt SHALL list `ontology_list_actions`, `ontology_action_*` dynamic tools, and governance tools (`ontology_promote_type`, `ontology_promote_predicate`, `ontology_schema_health`, `ontology_type_usage`).

#### Scenario: Ontologist identity prompt documents ACL and tool surface
- **WHEN** the ontologist identity prompt is loaded
- **THEN** it MUST describe the ontologist role, ACL caveats, dynamic ontology actions, and governance tools

### Requirement: Agent count documentation sync
All documentation and spec files referencing "7 built-in agents" SHALL be updated to "8 built-in agents" with ontologist included in the agent list.

#### Scenario: Ontologist increases documented built-in agent count
- **WHEN** documentation lists the built-in agents
- **THEN** it MUST describe 8 built-in agents and include the ontologist

### Requirement: Ontology governance and action tools keep actionable wrapper parameter guards

Ontology governance and dynamic action tools SHALL reject missing required wrapper inputs with actionable parameter errors before downstream ontology operations begin.

#### Scenario: Ontology governance and action tools reject missing required inputs
- **WHEN** `ontology_promote_type`, `ontology_promote_predicate`, `ontology_type_usage`, or any `ontology_action_*` tool is invoked without one of its declared required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not proceed into downstream ontology service execution

