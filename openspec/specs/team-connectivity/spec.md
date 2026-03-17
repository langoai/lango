# Spec: Team Connectivity

## Purpose

Wires P2P Team Coordinator with Escrow Engine, Budget Engine, and Workspace Manager through event-driven bridges. Enables agents to form teams, delegate tasks, and auto-settle via escrow without direct cross-package imports.

## Requirements

### R1: Team Event Publishing

All team lifecycle transitions MUST publish corresponding events via EventBus.

**Scenarios:**
- FormTeam() publishes TeamFormedEvent with teamID, name, goal, leaderDID, member count
- FormTeam() publishes TeamMemberJoinedEvent for each member (existing)
- DelegateTask() publishes TeamTaskDelegatedEvent before dispatch with worker count
- DelegateTask() publishes TeamTaskCompletedEvent after wg.Wait() with success/fail counts and average duration
- CollectResults() publishes TeamConflictDetectedEvent when resolver fails and >1 successful members
- DisbandTeam() publishes TeamMemberLeftEvent for each member (existing)
- DisbandTeam() publishes TeamDisbandedEvent with teamID and reason

### R2: Team Agent Tools

Five agent-invocable tools MUST be registered under the "p2p" catalog category.

**Scenarios:**
- `team_form` creates a team with UUID, selecting agents by capability
- `team_delegate` sends tool invocation to all workers and resolves conflicts
- `team_status` returns team details including members, budget, trust scores
- `team_list` returns all active teams with summary info
- `team_disband` disbands a team by ID

### R3: Team Protocol Handlers

All 5 team message types MUST be routed in the protocol handler switch.

**Scenarios:**
- `team_invite` → TeamRouter.OnInvite handler
- `team_accept` → TeamRouter.OnAccept handler
- `team_task` → TeamRouter.OnTask handler (executes tool locally)
- `team_result` → TeamRouter.OnResult handler
- `team_disband` → TeamRouter.OnDisband handler (calls coord.DisbandTeam)
- Unknown team types return error response

### R4: Remote Agent Team Methods

P2PRemoteAgent MUST expose 3 team-related methods.

**Scenarios:**
- SendTeamInvite() opens stream, sends RequestTeamInvite, returns Response
- SendTeamTask() opens stream, sends RequestTeamTask with deadline, returns Response
- SendTeamDisband() opens stream, sends RequestTeamDisband, returns Response

### R5: Team-Escrow Bridge

Team events MUST auto-manage escrow lifecycle when team has budget > 0.

**Scenarios:**
- TeamFormedEvent → creates escrow with per-worker milestones (budget split equally)
- TeamTaskCompletedEvent → completes next pending milestone with evidence
- TeamDisbandedEvent + all milestones done → releases escrow
- TeamDisbandedEvent + incomplete milestones → disputes then refunds escrow
- Team with zero budget → no escrow created (silently skipped)

### R6: Team-Budget Bridge

Team events MUST auto-manage budget allocation and spend tracking.

**Scenarios:**
- TeamFormedEvent → allocates budget, publishes TeamPaymentAgreedEvent per worker
- TeamTaskDelegatedEvent → reserves estimated cost (0.1 USDC per worker)
- TeamTaskCompletedEvent → records actual cost based on successful invocations
- Budget reservation auto-releases after 5-minute timeout

### R7: Workspace-Team Bridge

Team events MUST auto-manage workspace lifecycle.

**Scenarios:**
- TeamFormedEvent → creates workspace named "team-{teamID[:8]}"
- TeamTaskCompletedEvent → records contribution in workspace tracker
- TeamDisbandedEvent → unsubscribes gossip, cleans up mapping

### R8: Convenience Tools

Two high-level workflow tools MUST combine team + escrow + budget operations.

**Scenarios:**
- `team_form_with_budget` creates team + escrow + budget in single call
- `team_form_with_budget` supports explicit milestones or auto-splits among workers
- `team_complete_milestone` marks milestone complete and auto-releases if all done

### R9: App Wiring

All bridges, tools, and handlers MUST be wired in app.go and wiring_p2p.go.

**Scenarios:**
- Team tools registered after P2P tools in catalog
- Team-escrow bridge wired when both coordinator and escrow engine exist
- Team-budget bridge wired when both coordinator and budget engine exist
- Workspace-team bridge wired when workspace components exist
- Team protocol handler wired after coordinator creation
- Convenience tools registered when escrow engine exists
