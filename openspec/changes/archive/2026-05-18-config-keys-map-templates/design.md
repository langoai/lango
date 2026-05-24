## Design

`collectKeys` remains the single source for `lango config keys` and config-path suggestions. When it encounters a map with string keys, it should recurse into the map value type using a placeholder segment instead of treating the map as a terminal leaf.

## Placeholder Rules

- Struct-valued maps use `<name>` as the dynamic segment, then expose the value struct's leaf fields.
- String-valued maps use `<key>` as the dynamic segment because the dynamic segment itself is the leaf key.
- Dynamic map templates expose only leaf types accepted by `config set`; unsupported leaves such as `time.Duration` fields are omitted.
- Unsupported non-string-key maps remain omitted from discovery.

## Examples

- `providers.<name>.type`
- `providers.<name>.apiKey`
- `providers.<name>.baseUrl`
- `mcp.servers.<name>.env.<key>`
- `mcp.servers.<name>.headers.<key>`
- `auth.providers.<name>.clientSecret`

## Verification

Add focused unit coverage for `collectKeys` and command output, then run the full Go build/test and strict OpenSpec validation.
