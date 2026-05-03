# Multi-Agent Coworking Wave 5

## Why

Mission Control can already show durable missions, transient proposals, live decisions, and operator loops. What it still does not show clearly is how multiple local agents are collaborating inside one mission: who handed off to whom, who is currently involved, what is blocked on approval or teammate work, and whether budget or recovery pressure is building inside that mission.

Wave 5 exists to make local coworking visible without pretending that collaboration has become a new durable domain model or that external P2P team UX is already solved.

## What Changes

This change introduces the first mission-linked local coworking slice:

- add a collaboration projection attached to durable missions rather than a new durable collaboration table
- surface mission-linked local participants, handoffs, blocked state, budget pressure, and recovery hints
- require strict mission-linked attribution before session-level runtime signals can appear on a mission row
- keep external P2P team data secondary in the first slice

## Scope Guardrails

Wave 5 is intentionally narrow:

- first slice is mission-linked **local** coworking only
- collaboration state is projected, not durably persisted as a new work model
- session-level delegation, budget, and recovery signals must not be over-attributed to missions
- external P2P team UX remains secondary
- no cockpit controls for team formation, role editing, or conflict resolution in this slice

## User Impact

Mission Control becomes collaboration-aware. Users can tell which local agents are working on a mission, where the recent handoff happened, whether work is blocked on approval or teammate response, and whether recovery or budget pressure is part of the current mission context.
