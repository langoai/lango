## 1. Contract Sync

- [x] 1.1 Update `payment-execution-gating` to distinguish validation errors from deny outcomes.
- [x] 1.2 Update `payment-tools` coverage for `payment_send` missing transaction receipt ids.
- [x] 1.3 Sync public payment/security docs with the implemented validation-vs-deny split.

## 2. Verification

- [ ] 2.1 Run `openspec archive payment-execution-validation-contract-sync -y`.
- [ ] 2.2 Run `openspec validate --specs`.

## 3. Spec Sync

- [x] 3.1 Update `payment-execution-gating`, `payment-tools`, and `downstream-docs-sync` delta specs.
