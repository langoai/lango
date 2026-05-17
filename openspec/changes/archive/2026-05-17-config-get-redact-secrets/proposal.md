## Summary

Redact sensitive values from `lango config get` output by default, with an explicit `--show-secrets` override for intentional secret inspection.

## Motivation

`lango config set` now avoids leaking secrets in confirmations and supports `--from-env`, but `lango config get providers.openai.apiKey` still exposes the stored secret. Object and JSON reads such as `lango config get providers --output json` can also expose nested provider keys, MCP headers, tokens, or PINs. A production CLI should make safe output the default and require an explicit opt-in to display secrets.

## Scope

- Redact sensitive scalar paths in `config get` plain and JSON output by default.
- Redact nested sensitive fields in object/map outputs while preserving valid JSON shape.
- Add `--show-secrets` to intentionally bypass redaction.
- Reuse the same path sensitivity rules as `config set` to avoid drift.
- Update public CLI docs and OpenSpec coverage.

## Non-Goals

- Change `lango config export`, which is explicitly a plaintext backup/export command.
- Add a secret manager or encrypted display workflow.
- Change config persistence or validation.
