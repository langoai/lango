## Why

`lango config get` can read map-backed configuration paths such as `providers.openai.apiKey`, but `lango config set` stops when it reaches a map. This makes documented provider and MCP configuration paths effectively read-only from the CLI and pushes users back to manual JSON editing.

## What Changes

- Allow `lango config set` to traverse map-backed dot paths.
- Allow missing map entries to be created when the remaining path can be applied to the map value type.
- Support setting scalar fields inside map values, including provider and auth provider entries.
- Support setting string values in `map[string]string` leaves, including MCP env/header maps and pricing/custom-pattern maps.
- Preserve existing invalid-path behavior: invalid paths fail before save and continue to include key-discovery help where applicable.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `config-cli-commands`: `config set` can update documented map-backed dot paths without requiring manual profile edits.

## Impact

- Affected code: `internal/cli/configcmd/getset.go`, `internal/cli/configcmd/getset_test.go`.
- Affected user docs: config CLI examples may include concrete map-backed `config set` examples.
- No storage, bootstrap, encryption, or provider runtime interface changes.
