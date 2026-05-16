# Spec: Team Connectivity

## Purpose

Wires P2P Team Coordinator with Escrow Engine, Budget Engine, and Workspace Manager through event-driven bridges. Enables agents to form teams, delegate tasks, and auto-settle via escrow without direct cross-package imports.
## Requirements

### R1: Team Event Publishing

All team lifecycle transitions MUST publish corresponding events via EventBus.

#### Scenario: Team lifecycle emits the expected event stream
- **WHEN** a team is formed, delegated work, completes work, hits a conflict, or disbands
- **THEN** the coordinator MUST publish the corresponding team lifecycle events through EventBus

### R2: Team Agent Tools

Five agent-invocable tools MUST be registered under the "p2p" catalog category.

#### Scenario: Team tools cover core coordination actions
- **WHEN** the agent inspects the registered team tools
- **THEN** the `team_form`, `team_delegate`, `team_status`, `team_list`, and `team_disband` tools MUST cover the documented coordination actions

### R3: Team Protocol Handlers

All 5 team message types MUST be routed in the protocol handler switch.

#### Scenario: Team protocol messages route to the correct handlers
- **WHEN** team protocol messages are received
- **THEN** invite, accept, task, result, and disband messages MUST route to the matching handler
- **AND** unknown team message types MUST return an error response

### R4: Remote Agent Team Methods

P2PRemoteAgent MUST expose 3 team-related methods.

#### Scenario: Remote agent exposes invite, task, and disband methods
- **WHEN** a caller uses team methods on `P2PRemoteAgent`
- **THEN** invite, task, and disband requests MUST open streams, send the proper request type, and return responses

### R5: Team-Escrow Bridge

Team events MUST auto-manage escrow lifecycle when team has budget > 0.

#### Scenario: Escrow bridge follows team lifecycle
- **WHEN** budgeted team lifecycle events are published
- **THEN** the escrow bridge MUST create, advance, release, dispute, or skip escrow according to the documented team state

### R6: Team-Budget Bridge

Team events MUST auto-manage budget allocation and spend tracking.

#### Scenario: Budget bridge allocates, reserves, and records spend
- **WHEN** team formation, delegation, completion, and timeout events occur
- **THEN** the budget bridge MUST allocate budgets, reserve estimated cost, record actual spend, and auto-release stale reservations

### R7: Workspace-Team Bridge

Team events MUST auto-manage workspace lifecycle.

#### Scenario: Workspace bridge follows team lifecycle
- **WHEN** team lifecycle events are published
- **THEN** the workspace bridge MUST create the workspace, record contributions, and clean up mappings and gossip subscriptions on disband

### R8: Convenience Tools

Two high-level workflow tools MUST combine team + escrow + budget operations.

#### Scenario: Convenience tools combine team, budget, and escrow flows
- **WHEN** an operator uses the high-level team workflow tools
- **THEN** `team_form_with_budget` and `team_complete_milestone` MUST compose the documented team, budget, and escrow operations

### R9: App Wiring

All bridges, tools, and handlers MUST be wired in app.go and wiring_p2p.go.

#### Scenario: App wiring connects team bridges and handlers
- **WHEN** the P2P application wiring completes
- **THEN** team tools, bridges, handlers, and convenience tools MUST be registered in the documented conditions and order

### Requirement: Team workflow tools keep actionable wrapper parameter guards

Team workflow tools SHALL reject missing required wrapper inputs with actionable parameter errors before workflow orchestration begins.

#### Scenario: Team workflow tools reject missing required inputs
- **WHEN** `team_form`, `team_form_with_budget`, or `team_complete_milestone` is invoked without one of its declared required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not proceed into downstream coordination, escrow, or budget operations

