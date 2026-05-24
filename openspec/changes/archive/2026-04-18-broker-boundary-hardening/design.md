## Context

The storage facade already owns many domain factories, but some app and CLI paths still bypass it through generic handle access. The safest hardening step is to keep the underlying ent/sql wiring internal while exposing only capability-focused readers and constructors needed by production code.

## Goals / Non-Goals

**Goals:**
- Remove production app/CLI dependence on `Facade.EntClient()` and `Facade.RawDB()`.
- Replace raw access with narrow readers/factories for learning, inquiries, reputation, workflow, payment, observability alerts, and ontology dependencies.
- Keep bootstrap close semantics and test scaffolding working.

**Non-Goals:**
- Rewriting every bootstrap/test helper to avoid ent clients.
- Moving FTS/index bootstrap off the shared SQL handle in this change.
- Changing user-visible command output semantics.

## Decisions

- `Facade` keeps internal `client`/`rawDB` fields but no longer re-exports generic production accessors.
- Production CLI commands use storage readers returning domain/DTO records rather than querying Ent directly.
- Production app wiring uses prebuilt domain factories/dependency bundles (`OntologyDeps`, `ReputationStore`, `WorkflowStateStore`, `Alerts`, etc.) instead of reconstructing stores from `*ent.Client`.
- Any remaining raw DB exposure is restricted to narrow transitional methods such as FTS bootstrap, not generic app/CLI query access.

## Risks / Trade-offs

- The facade grows wider because it now owns more domain-specific constructors. This is acceptable because it hardens the boundary without a full broker rewrite.
- Some internal test harnesses still use `WithEntClient`; that is acceptable because the change targets production boundary enforcement first.
