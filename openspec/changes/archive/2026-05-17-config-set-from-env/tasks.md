## 1. Tests

- [x] 1.1 Add failing tests for `config set <path> --from-env <ENV>` saving the env value and redacting sensitive output.
- [x] 1.2 Add failing tests proving `--from-env` rejects missing env variables before loading or saving config.
- [x] 1.3 Add failing tests proving `--from-env` cannot be combined with a positional value.
- [x] 1.4 Add or preserve tests proving normal positional `config set` behavior still works.

## 2. Implementation

- [x] 2.1 Add `--from-env` flag and custom argument validation for positional versus env-sourced values.
- [x] 2.2 Resolve environment values with `os.LookupEnv` before config bootstrap.
- [x] 2.3 Reuse existing set/save/confirmation behavior after value resolution.

## 3. Documentation And Verification

- [x] 3.1 Update config CLI docs with `--from-env` usage and examples.
- [x] 3.2 Run focused config command tests, full Go build/tests, and strict OpenSpec validation.
- [x] 3.3 Run subagent-driven spec and code-quality review and address required findings.
- [x] 3.4 Sync specs, archive the OpenSpec change, and commit this scoped unit.
