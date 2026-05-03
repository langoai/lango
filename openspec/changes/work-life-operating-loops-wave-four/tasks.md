## 1. Loop Projection Domain

- [ ] 1.1 Define `LoopView`, `AgendaView`, loop-kind, and loop-status types for the first Wave 4 slice
- [ ] 1.2 Implement deterministic loop projection from durable missions, pending inquiries, dead-letter backlog, cron-job schedule state, and deterministic follow-up predicates
- [ ] 1.3 Add projection tests for session scoping, deterministic ordering, source coverage, and absence of fabricated scheduled loops

## 2. App Wiring

- [ ] 2.1 Expose only the loop-relevant readers Mission Control needs on the app boundary
- [ ] 2.2 Keep unsupported calendar, inbox, external task, and workflow-run loop sources absent in the first slice
- [ ] 2.3 Add wiring tests proving loop readers are present only when real sources exist

## 3. Mission Control Surface

- [ ] 3.1 Extend cockpit deps and projector inputs for loop and agenda projection
- [ ] 3.2 Add a compact loop / agenda surface to Mission Control without replacing the durable mission board
- [ ] 3.3 Add Mission Control tests for deterministic agenda ordering, unmatched runtime visibility, follow-up predicates, and narrow-terminal rendering

## 4. Public Truth And Verification

- [ ] 4.1 Audit landed Wave 4 behavior before updating public docs
- [ ] 4.2 Update README and cockpit docs to describe the real first-slice loop sources and explicit non-goals
- [ ] 4.3 Run docs build, OpenSpec validation/verification, and final Wave 4 review before archive
