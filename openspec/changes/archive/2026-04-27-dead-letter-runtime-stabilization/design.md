## Design Summary

This stabilization batch patches six narrow runtime and wiring gaps without changing the broader product surface:

- full dead-letter filter forwarding through the cockpit shell adapter
- default operator-principal injection for replay surfaces
- panic recovery in background task execution
- per-peer serialization of reputation updates
- `NaN` clamping in trust/reputation scoring
- dispatch-family classifier unification between CLI and cockpit

The batch is intentionally additive and scoped:

- no new runtime policy model
- no new operator commands
- no new domain state

Docs and docs-only OpenSpec are aligned only where the stabilized behavior is now user-visible or contract-relevant.
