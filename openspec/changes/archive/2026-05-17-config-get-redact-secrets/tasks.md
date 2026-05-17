## 1. Tests

- [x] 1.1 Add failing tests proving sensitive scalar `config get` plain and JSON output is redacted by default.
- [x] 1.2 Add failing tests proving `--show-secrets` returns raw sensitive scalar values.
- [x] 1.3 Add failing tests proving object/map JSON output redacts nested sensitive fields while preserving non-sensitive fields and valid JSON.
- [x] 1.4 Add failing tests proving redaction does not mutate the loaded config object and non-sensitive reads remain unchanged.

## 2. Implementation

- [x] 2.1 Add `--show-secrets` to `config get`.
- [x] 2.2 Apply path-based redaction before `printValue` unless `--show-secrets` is set.
- [x] 2.3 Implement recursive redaction for structs, maps, pointers, interfaces, and slices without mutating source config.

## 3. Documentation And Verification

- [x] 3.1 Update config CLI docs and docs guard for redacted `config get` output and `--show-secrets`.
- [x] 3.2 Run focused config command tests, full Go build/tests, and strict OpenSpec validation.
- [x] 3.3 Run subagent-driven spec and code-quality review and address required findings.
- [x] 3.4 Sync specs, archive the OpenSpec change, and commit this scoped unit.
