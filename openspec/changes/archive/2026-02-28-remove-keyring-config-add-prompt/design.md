## Context

The `security.keyring.enabled` config flag was introduced with OS keyring support but is never consulted at runtime — `bootstrap.go` uses `keyring.IsAvailable()` auto-detection exclusively. The flag adds unnecessary config surface (struct, defaults, TUI form, menu entry, state handler) without providing value. Additionally, users who enter a passphrase interactively must separately run `lango security keyring store` to persist it, which is a poor UX.

## Goals / Non-Goals

**Goals:**
- Remove `security.keyring.enabled` config flag and all associated UI/config plumbing
- Add an automatic keyring storage prompt in `bootstrap.go` after interactive passphrase entry
- Maintain backward compatibility — existing keyring CLI commands (`status`, `store`, `clear`) remain unchanged

**Non-Goals:**
- Changing the keyring Provider interface or availability detection logic
- Modifying the passphrase acquisition priority chain order
- Adding keyring support for non-interactive environments

## Decisions

1. **Remove config flag entirely rather than deprecate**: The flag was never functional (bootstrap ignores it). No deprecation cycle needed since removing it has zero runtime behavior change.

2. **Prompt placement — after `passphrase.Acquire()`, before database open**: The passphrase is needed before the DB opens, and the prompt is a lightweight stdin read. Placing it here ensures the user sees the prompt at the natural point in the startup flow.

3. **Use existing `prompt.Confirm()` for the prompt**: Reuses the `internal/cli/prompt` package which already handles terminal I/O consistently.

4. **Non-fatal keyring store failure**: If `krProvider.Set()` fails, emit a stderr warning and continue. The passphrase was already acquired successfully so startup should not be blocked.

## Risks / Trade-offs

- [Config schema change] Users with `security.keyring.enabled` in their config.json will see an unknown field warning from Viper. → Viper silently ignores unknown keys by default, so no impact.
- [Prompt in non-interactive context] If stdin is redirected but `passphrase.Acquire()` somehow returns `SourceInteractive`. → This cannot happen because `Acquire()` only returns `SourceInteractive` when `term.IsTerminal(stdin)` is true, which guarantees `prompt.Confirm()` will also work.
