## Why

`lango config set` already supports dynamic map-backed paths such as provider IDs, MCP server names, environment variable names, and OIDC provider IDs. `lango config keys` does not expose those runnable path templates, so users get incomplete discovery guidance for valid paths.

## What Changes

- List map-backed config paths as concrete templates using placeholders such as `<name>` and `<key>`.
- Cover provider, MCP server env/header, auth provider, and other map-backed config leaves.
- Add regression coverage so `config keys` stays aligned with `config set` map-path support.

## Impact

- CLI-only behavior change for `lango config keys`.
- No config storage or profile format changes.
- Documentation remains accurate because existing public docs already describe these dynamic config paths.
