## 1. Tests

- [x] 1.1 Add failing tests for creating and updating provider map-backed paths.
- [x] 1.2 Add failing tests for creating nested `map[string]string` values such as MCP server env entries.
- [x] 1.3 Add failing tests proving invalid map-backed paths fail before save.

## 2. Implementation

- [x] 2.1 Refactor `setConfigPath` traversal so map values can be updated and written back.
- [x] 2.2 Support creation of missing maps and type-compatible map entries.
- [x] 2.3 Preserve existing scalar parsing, pointer handling, and invalid-path discovery behavior.

## 3. Documentation And Verification

- [x] 3.1 Update config CLI documentation with a concrete map-backed `config set` example.
- [x] 3.2 Run focused config command tests, full Go build/tests, and strict OpenSpec validation.
- [x] 3.3 Run subagent-driven review and address required findings.
- [x] 3.4 Archive the OpenSpec change and commit this scoped unit.
