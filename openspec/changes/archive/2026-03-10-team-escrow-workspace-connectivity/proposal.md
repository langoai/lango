# Proposal: P2P Team-Escrow-Workspace Connectivity

## Problem

Lango has fully functional but disconnected subsystems: Team Coordinator, Escrow Engine, Budget Engine, Workspace Manager, and Settlement Service. Agents cannot form teams, delegate tasks, or auto-settle via escrow because:

1. No team tools registered — coordinator exists but no `buildTeamTools()` function
2. Only 2/9 team events published — TeamMemberJoined/Left only
3. No team protocol handlers — 5 message types defined but not routed
4. No team→escrow bridge — no auto-escrow on team formation
5. No team→budget bridge — Team.Budget field unused
6. No workspace→team bridge — workspace contributions don't map to team progress

## Solution

Wire all subsystems through the existing EventBus pattern with event-driven bridges:

- Publish all team lifecycle events (formed, disbanded, delegated, completed, conflict)
- Register team agent tools for AI-driven coordination
- Add team protocol handlers for P2P message routing
- Create event-driven bridges: team→escrow, team→budget, workspace→team
- Add convenience tools for combined team+escrow+budget operations

## Scope

- 10 work units across 3 implementation waves
- ~1500 LOC total (bridges, tools, handlers, tests)
- No new dependencies or breaking changes
- All bridges use existing EventBus SubscribeTyped pattern

## Non-Goals

- On-chain smart contract integration (existing settler handles this)
- New UI/CLI commands for team management (tools are agent-invocable)
- Workspace P2P gossip protocol changes
