---
title: Exec Safety
---

# Exec Safety

## Overview

The exec safety system provides policy-based command evaluation for all shell commands executed by the agent. It intercepts `exec` and `exec_bg` tool calls, evaluates them through a multi-stage pipeline, and returns a verdict: **allow**, **observe**, or **block**.

This runs as a toolchain middleware, applied before the approval system.

## Architecture

```
exec/exec_bg tool call
  │
  ├─ 1. Shell wrapper unwrap (sh -c, bash -c, env sh -c → inner command)
  ├─ 2. Env prefix stripping (VAR=val cmd → cmd)
  ├─ 3. Lango CLI / skill-import classification → Block
  ├─ 4. CommandGuard checks (kill verbs, protected paths) → Block
  ├─ 5. xargs/find-exec inner verb extraction → Block or Observe
  ├─ 6. Catastrophic pattern match → Block
  ├─ 7. Opaque pattern detection → Observe
  └─ 8. Pass → Allow
```

## Verdict System

| Verdict | Action | Description |
|---------|--------|-------------|
| **Allow** | Execute normally | Command is clearly safe |
| **Observe** | Execute + log + publish event | Command has opaque elements that prevent full static analysis |
| **Block** | Reject execution | Command is dangerous or matches a blocked pattern |

## Reason Codes

Each policy decision includes a machine-readable reason code:

| Reason Code | Verdict | Description |
|------------|---------|-------------|
| `kill_verb` | Block | Process management commands (`kill`, `pkill`, `killall`) |
| `protected_path` | Block | Access to protected paths (config DB, encrypted stores) |
| `lango_cli` | Block | Attempts to invoke the Lango CLI itself |
| `skill_import` | Block | Attempts to import skills via shell |
| `catastrophic_pattern` | Block | Matches a catastrophic safety pattern from configuration |
| `cmd_substitution` | Observe | Contains `$(...)` or backtick command substitution |
| `unsafe_var_expansion` | Observe | Contains `$VAR` or `${VAR}` where VAR is not in the safe set |
| `eval_verb` | Observe | Command verb is `eval` |
| `encoded_pipe` | Observe | Base64 decode piped to shell/eval |
| `heredoc` | Observe | Contains heredoc syntax |
| `process_substitution` | Observe | Contains process substitution `<(...)` or `>(...)` |
| `grouped_subshell` | Observe | Contains grouped subshell `(...)` or `{...}` |
| `shell_function` | Observe | Contains shell function definition |
| `xargs_inner` | Observe/Block | Contains `xargs` with inner command (blocked if inner verb is dangerous) |
| `find_exec_inner` | Observe/Block | Contains `find -exec` with inner command (blocked if inner verb is dangerous) |

## Shell Wrapper Unwrapping

The evaluator detects and recursively unwraps shell wrapper commands to evaluate the actual inner command:

**Supported wrappers:**

- `sh -c "cmd"`, `bash -c 'cmd'`, `zsh -c cmd`, `dash -c cmd`
- `/bin/sh -c`, `/usr/bin/bash -c` (absolute paths)
- `sh -lc`, `sh -ic` (login/interactive flags with -c)
- `env sh -c`, `/usr/bin/env bash -c` (env prefix)
- Nested wrappers up to 5 levels deep

**Parsing:** Uses AST-based parsing via `mvdan.cc/sh/v3/syntax` with a string-based fallback for commands that fail AST parsing.

## Opaque Pattern Detection

Commands that cannot be fully analyzed statically are given an **observe** verdict. The detector checks for:

1. **Command substitution** -- `$(...)` or backtick execution
2. **Unsafe variable expansion** -- `${VAR}` or `$VAR` where VAR is not in the safe set
3. **Eval verb** -- direct `eval` invocation
4. **Encoded pipe** -- base64/decode piped to shell or eval

## Safe Environment Variables

The following variables are considered safe and do not trigger `unsafe_var_expansion`:

`HOME`, `PATH`, `USER`, `PWD`, `SHELL`, `TERM`, `LANG`, `LC_ALL`, `LC_CTYPE`, `TMPDIR`

## Catastrophic Pattern Check

Commands are checked against a configurable list of catastrophic patterns (case-insensitive). These patterns are blocked before any approval system. Configure via:

```json
{
  "hooks": {
    "blockedCommands": ["rm -rf /", "mkfs", "dd if="]
  }
}
```

The default blocked patterns from `DefaultBlockedPatterns` are always applied and merged with any user-configured patterns.

## Guard Middleware

The `WithPolicy` middleware integrates the policy evaluator into the tool execution chain:

```
Tool Call → WithPolicy middleware → Approval middleware → Tool Handler
                │
                ├─ Block → return BlockedResult (skip all downstream)
                ├─ Observe → publish event + log → continue to next
                └─ Allow → continue to next
```

Only `exec` and `exec_bg` tools are evaluated. All other tools pass through unchanged.

## Event Publishing

Policy decisions with Block or Observe verdicts are published as `PolicyDecisionEvent` on the event bus (when `hooks.eventPublishing` is enabled) and logged to the session audit trail.
