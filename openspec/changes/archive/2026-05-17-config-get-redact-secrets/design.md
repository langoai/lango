## Context

`config get` resolves a dot path and passes the resulting value directly to `printValue`. This is safe for ordinary scalar settings but unsafe for credential paths and for object reads containing nested credentials. The CLI already has deterministic path sensitivity helpers for `config set`, so `config get` can reuse those helpers before formatting output.

## Decisions

1. Redact by resolved config path, not by value contents.

   Values are arbitrary and cannot reliably identify secrets. Dot paths provide stable signal and match the existing `config set` behavior.

2. Redact recursively for object and map reads.

   For `config get providers --output json`, only nested sensitive leaves should be replaced with `<redacted>` while non-sensitive fields stay visible. This keeps JSON decodable and useful without leaking credentials.

3. Preserve raw reads behind `--show-secrets`.

   Operators sometimes need to inspect a stored value. `--show-secrets` makes that intent explicit and keeps existing troubleshooting capability available.

4. Keep export behavior unchanged.

   `config export` is documented as plaintext backup output and already carries a security warning. This change affects `config get` only.

## Risks / Trade-offs

- Some future credential names may not be redacted until the shared path sensitivity helper is extended.
- Some non-secret values may be redacted if their path uses a credential-like name. That is safer than leaking secrets and does not affect stored data.
- Recursive redaction needs reflection care to avoid mutating the loaded config object; tests should verify redaction output does not alter source values.
