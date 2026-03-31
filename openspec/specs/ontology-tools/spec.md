## ADDED Requirements

### Requirement: Ontology surface tools
The system SHALL provide 13 static agent-facing tools with `ontology_` prefix plus dynamic `ontology_action_{name}` tools generated from the ActionRegistry and an `ontology_list_actions` tool. `BuildTools` SHALL accept `(svc OntologyService, reg *ActionRegistry)` — when reg is nil, only static tools are returned. Read-only tools SHALL use SafetyLevelSafe; mutation tools SHALL use SafetyLevelModerate.

### Requirement: Ontologist agent routing
The system SHALL define an `ontologist` AgentSpec with Prefixes `["ontology_"]`. All `ontology_` prefixed tools SHALL be routed to the ontologist agent via PartitionTools.

### Requirement: Ontology tools registration
When `ontology.enabled` is true and OntologyService is initialized, the system SHALL register ontology tools via CatalogEntry in the intelligence module.

### Requirement: JSON import tool
`ontology_import_json` SHALL accept a `data` parameter (JSON string) containing entities with id, type, properties, and optional relations. Each entity SHALL be validated via SetEntityProperty (type+property schema) and relations via AssertFact (predicate validation).

### Requirement: CSV import tool
`ontology_import_csv` SHALL accept `data` (CSV string) and `type` (ObjectType name). The first row SHALL be treated as property name headers. Each subsequent row creates an entity with the given type and column values as properties.

### Requirement: MCP result import tool
`ontology_from_mcp` SHALL accept `tool_name`, `result_json` (JSON string), `entity_type`, and `predicate`. The handler SHALL decode the JSON, create entity properties, and assert a fact linking the entity to the tool via the specified predicate. Explicit mapping only — no automatic type inference.

### Requirement: Ontologist identity prompt
The system SHALL provide a `prompts/agents/ontologist/IDENTITY.md` file defining the ontologist agent's role, capabilities, and tool usage guidelines. The identity prompt SHALL include a note that ontology operations may be restricted by ACL permissions based on the calling agent's role. The identity prompt SHALL list `ontology_list_actions` and `ontology_action_*` dynamic tools.

### Requirement: Agent count documentation sync
All documentation and spec files referencing "7 built-in agents" SHALL be updated to "8 built-in agents" with ontologist included in the agent list.
