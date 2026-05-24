## Why

The executable provenance guard already checks the README quick reference against the current concrete command list, but the main `docs-only` spec still describes the family with broad shorthand like `status/checkpoint/session/attribution/bundle`.

## What Changes

- expand the `docs-only` provenance requirement to the current explicit command list
- extend the existing provenance completeness guard so it also enforces the `docs-only` main spec wording

## Impact

- more truthful main spec wording for provenance commands
- one executable guard covering both README and the main docs-only spec
- less chance of broad shorthand drifting back into requirements
