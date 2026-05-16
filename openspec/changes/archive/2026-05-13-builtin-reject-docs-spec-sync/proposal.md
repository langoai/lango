## Why

Public multi-agent docs still claim that built-in teammates reject misrouted work with textual `[REJECT]` markers, but the current built-in prompt/runtime contract uses a short visible escalation summary instead. The main `agent-routing` spec also mixes the old reject wording with the newer built-in escalation rule, which makes the source of truth harder to follow.

## What Changes

- Update `README.md` so the built-in teammate runtime contract matches the current visible-escalation behavior.
- Clarify `agent-routing` so `[REJECT]` remains a compatibility safety net for legacy or remote sub-agent paths, while built-in teammate prompt overrides stay free of `[REJECT]`.
- Sync downstream docs/spec requirements to the same public contract.

## Impact

- `agent-routing`: built-in vs compatibility rejection semantics are easier to reason about.
- `downstream-docs-sync`: README no longer advertises a built-in behavior that production prompts explicitly forbid.
