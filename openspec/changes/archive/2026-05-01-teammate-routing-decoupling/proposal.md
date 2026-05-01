# Teammate Routing Decoupling

## Why

The current code no longer prefixes spawned teammate instructions with an advisory system string. Instead, it stores the raw user instruction and carries teammate intent through durable metadata such as `RequestedAgent` and the runtime allowlist.

The archived specs still describe advisory routing through an enriched prompt prefix, which no longer matches the implementation. The capability policy also now treats non-built-in teammates differently from built-in teammate types: custom teammate paths use the current allowlist as their effective ceiling rather than a built-in role ceiling.

## What Changes

- Reframe `agent_spawn` so advisory routing is carried by `RequestedAgent` metadata instead of an instruction prefix.
- Update the control-plane contract so the stored `Instruction` remains the raw user instruction.
- Clarify that built-in teammate types enforce a role maximum scope, while custom teammate paths have no built-in escalation path beyond the current spawn-time allowlist / projected allowlist.

## Impact

- Removes the remaining spec/code mismatch around enriched advisory prefixes.
- Makes custom teammate capability behavior explicit for future runtime work.
