# Design: Config Key Suggestions

## Approach

Keep the implementation inside `internal/cli/configcmd/getset.go`, where config key traversal already lives. Add a small unexported error-formatting helper that receives the full path, failed segment, successfully traversed prefix, and available keys from `collectKeys(reflect.TypeOf(config.Config{}), "")`.

When a struct field segment is missing in `resolveConfigPath` or `setConfigPath`, return an error with:

- the original path and missing segment,
- a `did you mean` list for close valid keys,
- a `list keys: lango config keys <prefix>` hint.

The suggestion algorithm should be deterministic, dependency-free, and conservative:

- Prefer valid keys that share the same parent prefix as the failed path.
- Fall back to full-key similarity when there are no same-prefix candidates.
- Rank by edit distance, then lexical order.
- Limit the output to three keys.

## Error Shape

Example:

```text
config path "agent.providr": field "providr" not found; did you mean: agent.provider; list keys: lango config keys agent
```

If there are no useful suggestions, the error should still include the discovery hint:

```text
config path "made.up.path": field "made" not found; list keys: lango config keys
```

## Testing

Add RED-first tests in `internal/cli/configcmd/getset_test.go` for:

- `config get agent.providr` includes `agent.provider` and `lango config keys agent`.
- `config set knowledge.enable false` includes `knowledge.enabled` and does not call save.
- Unknown top-level paths include a generic `lango config keys` hint.
- Valid `config get`, `config set`, and `config keys` behavior remains unchanged.

## Tradeoffs

This intentionally does not create a reusable global suggestion package. The feature is specific to config dot-paths and relies on config mapstructure tags, so keeping it in `configcmd` avoids a premature abstraction.
