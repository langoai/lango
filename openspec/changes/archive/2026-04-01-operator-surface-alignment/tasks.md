## 1. Observability Docs Update

- [x] 1.1 Add "Policy Metrics" section to `docs/features/observability.md` documenting `/metrics/policy` endpoint, `PolicyDecisionEvent` flow, and `RecordPolicyDecision` collector method
- [x] 1.2 Add "Policy Decision Audit Logging" subsection to the Audit Logging section in `docs/features/observability.md` documenting the `PolicyDecisionEvent` -> audit recorder flow
- [x] 1.3 Add "Recovery Decision Events" section to `docs/features/observability.md` documenting `RecoveryDecisionEvent`, exponential backoff formula, and per-error-class retry limits table
- [x] 1.4 Update the Gateway Endpoints summary table in `docs/features/observability.md` to include `/metrics/policy`

## 2. Provenance Docs Update

- [x] 2.1 Add "Config and Hook Provenance" section to `docs/features/provenance.md` documenting config fingerprint computation and hook registry snapshot in checkpoint metadata

## 3. CLI Metrics Docs Update

- [x] 3.1 Add `lango metrics policy` command reference section to `docs/cli/metrics.md` with usage, flags, and example output
