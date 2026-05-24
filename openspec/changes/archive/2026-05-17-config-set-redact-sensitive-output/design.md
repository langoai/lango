## Context

The `config set` command writes `Set <path> = <value>` after a successful save. That is useful feedback for simple scalar values, but it is unsafe for paths that carry credentials. The command already receives the raw path and value, so output masking can happen entirely in the CLI layer without changing persistence.

## Goals / Non-Goals

**Goals:**

- Prevent credential values from appearing in `config set` success output.
- Keep saved configuration values unchanged.
- Detect sensitive dynamic map leaves such as `mcp.servers.<name>.env.API_KEY`.
- Keep non-sensitive confirmations readable.

**Non-Goals:**

- Build a general secret scanner.
- Change how config values are saved, expanded, validated, or encrypted.
- Change command arguments or add new flags.

## Decisions

1. Redact by path, not by value.

   Values can be arbitrary strings and may not be distinguishable from non-secrets. The config path is the stable signal available before output.

2. Normalize path segments before matching sensitive names.

   Segment normalization should remove non-alphanumeric characters and lower-case the result so `apiKey`, `API_KEY`, and `api-key` are treated consistently.

3. Use conservative sensitive markers.

   A segment is sensitive when its normalized form ends with credential names such as `apikey`, contains authorization markers, contains markers such as `secret` or `password`, exactly names PIN fields, or ends with credential suffixes such as `credential`, `credentials`, `token`, `privatekey`, or `accesskey`. This covers known config credentials and dynamic header/env names while avoiding broad redaction of non-secret paths like `agent.maxTokens`, `p2p.keyDir`, `p2p.zkp.maxCredentialAge`, and `security.signer.keyId`.

## Risks / Trade-offs

- A future credential field with an unusual name may not be redacted. Mitigation: tests cover current credential patterns and the marker list is easy to extend.
- Some non-secret values may be redacted if their path uses a sensitive marker. Mitigation: confirmation masking is safer than leaking secrets and does not affect saved data.
