## Context

The secrets CLI already routes prompt output and result messaging through Cobra command streams, but `delete` still keeps its own prompt parsing path. The shared confirmation helper now covers the same semantics and is already used in neighboring config, security, extension, and payment commands.

## Goals / Non-Goals

**Goals:**
- Route `secrets delete` confirmation through the shared prompt helper
- Preserve the existing `--force` bypass and refusal guidance
- Extend regression coverage to the non-interactive guard path

**Non-Goals:**
- Changing prompt wording
- Refactoring `secrets set` or `secrets list`
- Changing secret storage business logic

## Decisions

Keep a command-local non-interactive guard, but delegate prompt formatting/parsing to `prompt.ConfirmIO(...)`.
Rationale: the guard is command-specific policy; the prompt interaction rules should remain centralized.

Generalize the non-interactive guard to any non-terminal `*os.File` input, not just `os.Stdin`.
Rationale: this matches the stronger contract already applied to `keyring clear` and `payment send`, and closes a wrapper/pipe inconsistency.

## Risks / Trade-offs

- [Risk] EOF denial now comes through the shared helper instead of the local parser. → Mitigation: keep the same `Aborted.` user-facing behavior in command-level tests.
- [Trade-off] A small local guard still remains. → Mitigation: limit it to the policy check and keep all prompt I/O semantics shared.
