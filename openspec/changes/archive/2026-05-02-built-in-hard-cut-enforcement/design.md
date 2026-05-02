# Design

## Problem

The built-in hard cut is incomplete while `BuildAgentTree()` continues to attach built-in teammate specs as ADK sub-agents. In that state, prompt guidance and runtime behavior disagree: the model is told to use `agent_spawn`, but `transfer_to_agent` still exists as a production tool surface.

## Decision

The production ADK tree keeps only:

- the root orchestrator
- remote A2A sub-agents
- explicit non-built-in custom specs

Built-in teammate specs stay in the routing table but do not become ADK sub-agents.

## Recovery Adjustment

Once zero built-in sub-agents becomes the normal steady state, hallucinated built-in targets must still trigger one correction attempt. The correction gate therefore stops using `len(SubAgents()) == 0` as a blanket short-circuit and only suppresses retries when there is neither a built-in target nor any remote/legacy sub-agent surface to retry against.
