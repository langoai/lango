## MODIFIED Requirements

### Requirement: GraphNode carries entity type information
GraphNode SHALL include a `NodeType string` field populated from the discovered triple's SubjectType metadata. AssembleSection SHALL format typed nodes as `**NodeType:ID**` when NodeType is non-empty.

#### Scenario: GraphNode populated with type from traversal
- **WHEN** graph expansion discovers a node from a triple with `SubjectType: "ErrorPattern"`
- **THEN** the resulting GraphNode has `NodeType: "ErrorPattern"`

#### Scenario: GraphNode from untyped triple
- **WHEN** graph expansion discovers a node from a triple with empty SubjectType
- **THEN** the resulting GraphNode has `NodeType: ""` and is formatted as `**ID**` (no type prefix)
