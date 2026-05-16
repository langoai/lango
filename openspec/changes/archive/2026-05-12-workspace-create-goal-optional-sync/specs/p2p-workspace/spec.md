## ADDED Requirements

### Requirement: Workspace creation keeps goal optional

The `p2p_workspace_create` operator tool SHALL require only a workspace name and SHALL treat the workspace goal as optional descriptive context.

#### Scenario: Workspace create accepts name without goal
- **WHEN** `p2p_workspace_create` is invoked with `name` and without `goal`
- **THEN** the tool SHALL create the workspace successfully
- **AND** the returned workspace payload SHALL keep an empty goal value
