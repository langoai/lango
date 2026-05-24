## 1. Tests

- [x] 1.1 Add failing tests proving sensitive `config set` provider output is redacted while the raw value is saved.
- [x] 1.2 Add failing tests proving sensitive dynamic map output is redacted while the raw value is saved.
- [x] 1.3 Add or preserve tests proving non-sensitive `config set` output remains readable.

## 2. Implementation

- [x] 2.1 Add a small path-based display helper for `config set` success output.
- [x] 2.2 Use the helper in `NewSetCmd` without changing saved config values.
- [x] 2.3 Keep sensitive path matching deterministic and covered by tests.

## 3. Documentation And Verification

- [x] 3.1 Update config CLI documentation to state sensitive success output is redacted.
- [x] 3.2 Run focused config command tests, full Go build/tests, and strict OpenSpec validation.
- [x] 3.3 Run subagent-driven review and address required findings.
- [x] 3.4 Sync specs, archive the OpenSpec change, and commit this scoped unit.
