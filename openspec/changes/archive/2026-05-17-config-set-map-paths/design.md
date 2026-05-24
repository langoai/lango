## Context

`internal/cli/configcmd/getset.go` uses reflection for `config get`, `config set`, and `config keys`. `resolveConfigPath` already traverses maps by consuming a path segment as the map key. `setConfigPath` only traverses structs, so it returns a non-struct path error as soon as it reaches a map.

The practical failure is user-facing: documented paths such as `providers.<id>.apiKey`, `auth.providers.<id>.clientId`, and `mcp.servers.<name>.env.<VAR>` cannot be set through the CLI even though users are told to use dot notation.

## Goals / Non-Goals

**Goals:**

- Make `config set` traverse maps consistently with `config get`.
- Create missing maps and map entries when the remaining path is compatible with the map value type.
- Preserve existing behavior for struct fields, pointer fields, scalar parsing, explicit-key persistence, cleanup, and invalid-path no-save semantics.
- Keep the implementation local to the config CLI reflection helper.

**Non-Goals:**

- Add provider-specific validation or health checks to `config set`.
- Add a new dedicated provider-management command.
- Add support for arbitrary JSON object values on the command line.
- Change encrypted profile storage or bootstrap behavior.

## Decisions

1. Recurse through set operations instead of trying to keep the current linear loop.

   Map values returned by `MapIndex` are not addressable. A recursive helper can copy the map value into an addressable value, apply the remainder of the path, then write the updated value back with `SetMapIndex`.

2. Create missing map entries from the map value type when the path continues.

   This makes `lango config set providers.openai.type openai` usable on a default profile where `Providers` starts empty. The created entry uses zero values except for the targeted field.

3. Treat the final segment of `map[string]string` as a leaf key.

   This supports documented nested maps such as MCP server env/header values and P2P tool prices without requiring JSON editing. Non-string map keys and complex final map values remain unsupported unless the path continues into a struct value.

4. Keep invalid missing struct fields actionable through existing discovery errors.

   Dynamic map keys cannot be fully enumerated by `config keys`, so missing map entries are created when type-compatible instead of producing suggestions for a key that cannot appear in static discovery output.

## Risks / Trade-offs

- Creating map entries from typos can persist unintended provider/server names. Mitigation: only create entries when the remaining path is valid for the value type, so `providers.opnai.unknown` still fails before save.
- Generic reflection can become hard to reason about. Mitigation: keep helper functions small and cover representative map-of-struct, nested map, pointer, and invalid-path cases.
- `config set providers.openai.apiKey ...` can create a provider entry without setting its `type`. Mitigation: this mirrors manual partial configuration and validation remains responsible for full profile correctness.
