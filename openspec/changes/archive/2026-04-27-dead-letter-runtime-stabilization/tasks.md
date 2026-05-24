## 1. Cockpit / retry wiring

- [x] 1.1 Forward all dead-letter filter fields through `cmd/lango` into the cockpit bridge
- [x] 1.2 Inject a default principal into CLI retry
- [x] 1.3 Inject a default principal into cockpit retry

## 2. Runtime hardening

- [x] 2.1 Recover from panics in background task execution
- [x] 2.2 Add or update background manager tests for panic recovery

## 3. Reputation hardening

- [x] 3.1 Serialize reputation updates per peer
- [x] 3.2 Add concurrency coverage for reputation updates
- [x] 3.3 Clamp `NaN` reputation scores

## 4. Family classifier consistency

- [x] 4.1 Reuse the shared dispatch-family classifier in cockpit summary aggregation

## 5. Docs / OpenSpec

- [x] 5.1 Update dead-letter docs for principal injection and cockpit filter forwarding
- [x] 5.2 Update docs-only OpenSpec requirements
