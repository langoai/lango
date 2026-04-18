## Context

Bootstrap currently depends on direct ent/sql access for config profile loading and session/recall wiring. Before parent DB open can be removed, broker-backed adapters for those bootstrap-facing capabilities need to exist.

## Goals / Non-Goals

**Goals:**
- Add broker RPCs for config profile store operations.
- Add broker-backed session store scaffolding for bootstrap/runtime wiring.
- Keep the repository buildable and testable while ownership migration proceeds incrementally.

**Non-Goals:**
- Completing the full parent DB-open removal in this change.
- Replacing every runtime domain store with broker-backed adapters yet.
- Finishing enforcement or raw-handle removal.

## Decisions

- Use broker-backed adapters that satisfy existing interfaces (`storage.ConfigProfileStore`, `session.Store`) where possible.
- Implement bootstrap-facing RPCs first so `phaseLoadProfile` and session wiring can move without redesigning all higher layers.
- Keep direct parent DB open in place until the broker-backed adapters are fully wired and validated.

## Risks / Trade-offs

- This is an intermediate ownership-migration step, so the true-ownership completion criteria are not met yet.
- RPC surface area grows, but it is necessary to decouple bootstrap from direct ent/sql handles.
