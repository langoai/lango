# Proposal: Config Key Suggestions

## Summary

Make `lango config get` and `lango config set` invalid dot-path errors actionable by showing nearby valid keys and the relevant `lango config keys <prefix>` discovery command.

## Problem

Invalid config paths currently fail with low-level reflection messages such as `field "providr" not found`. This tells the user what failed internally, but not how to recover. For a large configuration surface, typos like `agent.providr` or stale guesses like `knowledge.enable` should point users toward likely valid keys instead of forcing them to run a separate discovery command.

## Goals

- Keep config path resolution and mutation behavior unchanged for valid paths.
- Return actionable errors for invalid struct path segments used by `config get` and `config set`.
- Include up to a small number of deterministic nearby valid dot-path suggestions.
- Include a `lango config keys <prefix>` hint based on the valid prefix already traversed.
- Keep map lookup errors and type conversion errors unchanged unless they can be made more specific without hiding the original context.

## Non-Goals

- Do not introduce fuzzy matching dependencies.
- Do not change the list of valid config keys.
- Do not alter profile storage, bootstrap, or config validation behavior.
- Do not add shell completion or interactive prompting.
