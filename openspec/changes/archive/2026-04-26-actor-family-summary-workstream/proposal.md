## Why

Dead-letter operators can already inspect raw latest manual replay actor strings, but raw actors alone make it slow to spot whether the current backlog is mostly operator-driven, system-driven, or service-driven.

## What Changes

- document grouped actor-family summaries for `lango status dead-letter-summary`
- document the `By actor family` table section and `by_actor_family` JSON field
- document the cockpit dead-letter summary-strip `actor families:` line
- document the initial actor-family taxonomy: `operator`, `system`, `service`, and `unknown`
- preserve raw top latest manual replay actors alongside the grouped view
- sync public docs and docs-only OpenSpec requirements

## Impact

- operators get a faster backlog-level actor signal without losing the raw actor strings needed for diagnosis
- the CLI and cockpit summaries stay additive and backward compatible
- no Go control-plane behavior, retry semantics, backend read model, or filtering behavior changes are introduced by this docs archive
