## Context

The current runtime spreads DB ownership across bootstrap, app wiring, direct session stores, and bootstrap-free CLI helpers. A broker boundary must consolidate SQLite ownership before driver or payload protection changes can be made safely.

## Goals / Non-Goals

**Goals:**
- Make the broker the sole owner of SQLite open/migrate/init responsibilities.
- Replace direct DB exposure in bootstrap and app wiring with a storage facade.
- Remove status/doctor/settings/config profile bypass reads.

**Non-Goals:**
- Migrating graph persistence from BoltDB.
- Introducing payload encryption or modernc driver changes in this change.
- Adding auto-restart or complex broker process supervision.

## Decisions

- Use a persistent stdio JSON protocol with coarse-grained RPCs only.
- Keep graph.db outside broker scope for now, but protect it via sandbox/path policy.
- Fail closed on broker crash and do not auto-restart in v1.
- Compute protected paths from resolved runtime paths after config load, not from static defaults only.

## Risks / Trade-offs

- [Risk] Facade scope is broad and touches many DB-backed domains. → Mitigation: migrate by capability groups but keep bootstrap result API change atomic at the end.
- [Risk] Broker startup ordering can create config/profile load cycles. → Mitigation: explicitly make broker open/config load part of bootstrap.
