## MODIFIED Requirements

### Requirement: Ontologist identity prompt
The system SHALL provide a `prompts/agents/ontologist/IDENTITY.md` file defining the ontologist agent's role, capabilities, and tool usage guidelines. The identity prompt SHALL include a note that ontology operations may be restricted by ACL permissions based on the calling agent's role.

#### Scenario: Identity prompt mentions ACL
- **WHEN** ontologist agent identity is loaded
- **THEN** the prompt SHALL contain guidance that operations may be restricted by permissions and that write/admin operations require appropriate role assignment
