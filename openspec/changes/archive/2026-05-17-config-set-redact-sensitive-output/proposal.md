## Why

`lango config set` currently echoes the raw value in its success message. For credential paths such as provider API keys, auth client secrets, channel tokens, or MCP env/header credentials, that can expose secrets in terminal scrollback, logs, and agent transcripts.

## What Changes

- Redact sensitive values in `lango config set` success output.
- Preserve the actual saved value; only the confirmation text is masked.
- Keep non-sensitive `config set` confirmations unchanged.
- Document that sensitive paths are confirmed as `<redacted>`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `config-cli-commands`: `config set` success output must not reveal sensitive values.

## Impact

- Affected code: `internal/cli/configcmd/getset.go`, `internal/cli/configcmd/getset_test.go`.
- Affected docs: `docs/cli/config.md`.
- No storage, bootstrap, encryption, or config schema changes.
