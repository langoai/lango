# Design

## Advisory Routing

`RequestedAgent` is now the advisory routing signal for spawned teammates. The runtime may use that metadata when building teammate context, prompts, or routing decisions, but the stored `Instruction` remains the raw user instruction.

This keeps user intent, routing metadata, and runtime policy as separate concerns:

- `Instruction`: raw user task text
- `RequestedAgent`: advisory teammate identity
- `AllowedTools`: effective per-run ceiling for non-built-in teammate paths

## Built-In Versus Custom Teammates

Built-in teammate types continue to use role maximum scope as their ceiling. Custom or non-built-in teammate paths do not have that built-in ceiling, so runtime capability checks must treat the current allowlist as the effective ceiling.

That means:

- built-in teammates may request in-scope capability expansion
- custom teammates cannot escalate beyond the tools already present in `CurrentAllowed`

## Scope

This change is contract-only. It aligns specs with the current implementation and does not introduce a new runtime surface.
