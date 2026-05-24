## Context

`/mode` and `/cost` are first-class built-in slash commands in the runtime and are already visible in `/help` plus the cockpit feature docs. The only lagging artifact is the `lango cockpit` overview bullet list in `docs/cli/core.md`.

## Goals / Non-Goals

**Goals:**

- Keep the CLI overview command summary synchronized with the current slash-command surface.

**Non-Goals:**

- Change command behavior.
- Expand the CLI overview into a full command reference.

## Decisions

- Extend the existing slash-command bullet in place rather than adding a new subsection.
  - Rationale: the overview already has the right structure; it just needs the missing commands.

## Risks / Trade-offs

- [CLI overview line becomes slightly longer] → The added commands are high-value and still fit the overview style.
