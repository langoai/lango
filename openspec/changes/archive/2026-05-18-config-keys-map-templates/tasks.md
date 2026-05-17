## 1. Tests

- [x] 1.1 Add failing unit coverage that `collectKeys` includes dynamic map-backed config templates.
- [x] 1.2 Add failing command-output coverage that `lango config keys <prefix>` prints dynamic templates for provider, MCP, and auth maps.

## 2. Implementation

- [x] 2.1 Update config key collection to recurse into string-keyed map value types with deterministic placeholders.
- [x] 2.2 Keep existing static key discovery and invalid-path suggestions working.

## 3. OpenSpec And Verification

- [x] 3.1 Add delta specs for dynamic `config keys` templates and test coverage.
- [x] 3.2 Run focused tests, full Go build/tests, and strict OpenSpec validation.
- [x] 3.3 Run subagent-driven review and address required findings.
- [x] 3.4 Sync specs, archive the change, and commit the scoped unit.
