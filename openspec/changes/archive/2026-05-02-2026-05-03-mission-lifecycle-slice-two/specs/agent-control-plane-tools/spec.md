## MODIFIED Requirements

### Requirement: Task management tools provide CRUD on TaskEntry

Task management tools SHALL continue providing lightweight CRUD over `TaskEntry` in Slice 2. This change SHALL NOT promote `TaskEntry` into the durable mission checklist model or make task-tracking rows the authoritative durable mission truth.

#### Scenario: Task tracking remains operational rather than durable mission truth
- **WHEN** task management tools create, list, or update `TaskEntry` rows
- **THEN** those rows SHALL remain lightweight operational tracking records
- **AND** Slice 2 SHALL NOT require them to serve as the authoritative durable mission checklist model

## ADDED Requirements

### Requirement: Mission-aware execution linkage attaches at execution creation sites

When control-plane or mission-bound runtime work creates a new execution for an existing mission context, the system SHALL attach the durable mission-to-execution relationship at the execution creation site. `MissionExecutionLink` SHALL be the durable truth for that relationship rather than later inference from unrelated task-tracking records.

#### Scenario: Mission-bound spawned execution writes link at creation time
- **WHEN** mission-aware control-plane tooling creates a new execution for a mission
- **THEN** the application SHALL record the `MissionExecutionLink` as part of that execution creation flow
- **AND** the durable relationship SHALL reference the mission's `mission_id` plus the execution identity created by that flow

#### Scenario: Slice 2 does not retrofit all task tracking into mission linkage truth
- **WHEN** a `TaskEntry` exists without mission-aware execution linkage
- **THEN** Slice 2 SHALL NOT require the system to reconstruct durable mission ownership only by retrofitting all task-tracking records
- **AND** mission-execution linkage truth SHALL remain attached to execution creation sites
