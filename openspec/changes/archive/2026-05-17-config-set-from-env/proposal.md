## Summary

Add `lango config set --from-env <ENV>` so users can set configuration values from an environment variable instead of passing the raw value as a command argument.

## Motivation

`lango config set` now redacts sensitive values from success output, but users still need to type secrets directly into argv and shell history. That is unsafe for production setup and makes secure configuration harder than it should be. A first-class environment-variable input path lets users keep raw secrets out of command examples, terminal scrollback, and process argument listings while preserving the existing encrypted profile save flow.

## Scope

- Allow `lango config set <dot.path> --from-env <ENV>` as an alternative to `lango config set <dot.path> <value>`.
- Read the value with environment variable presence semantics, so an existing empty variable is saved as an empty string.
- Reject ambiguous usage such as combining `--from-env` with a positional value.
- Fail before loading or saving config when `--from-env` names an unset variable.
- Preserve existing redaction, explicit-key metadata, map-backed path support, cleanup, and command-output behavior.
- Update CLI documentation and OpenSpec coverage.

## Non-Goals

- Add stdin prompting or secret-manager integration.
- Change encrypted profile storage, env expansion during save, or config validation.
- Change `config get`, `config keys`, or profile-management commands.
