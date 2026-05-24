## Context

The existing `config set` command accepts exactly two positional arguments: a dot path and a value. This is convenient for scalar settings, but it pushes credential values into argv. The command already owns the CLI parsing boundary and save flow, so `--from-env` can be implemented entirely in `internal/cli/configcmd` without changing core config persistence.

## Decisions

1. Use `--from-env <ENV>` as an alternative value source.

   The command shape becomes `lango config set <dot.path> [value]`. If `--from-env` is empty, the existing two-argument behavior remains unchanged. If `--from-env` is present, the command requires exactly one positional argument: the config path.

2. Resolve the environment variable before bootstrapping config.

   Missing env variables should fail before passphrase/bootstrap work and before any save side effects. `os.LookupEnv` is required instead of `os.Getenv` so an explicitly set empty string remains a valid value.

3. Reject ambiguous input.

   `lango config set providers.openai.apiKey raw --from-env OPENAI_API_KEY` is invalid because it defeats the purpose of avoiding argv secrets and creates unclear precedence. The command should return an actionable error before loading config.

4. Reuse existing save/output paths.

   Once the value is resolved, the command should call `setConfigPath`, update explicit keys, save, and format output through the same logic as positional values. Sensitive paths still print `<redacted>`.

## Risks / Trade-offs

- Environment variables can still be exposed by a user's shell or process environment. This is materially safer than argv for common CLI usage, but not a replacement for a secret manager.
- `--from-env` with an empty but present variable can clear a value. That is intentional because `LookupEnv` distinguishes an explicit empty setting from a missing variable.
