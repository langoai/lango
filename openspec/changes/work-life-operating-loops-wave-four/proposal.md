# Work-Life Operating Loops Wave Four

## Why

Wave 3 added transient proactive proposals, but Mission Control still makes the operator infer higher-level recurring responsibility by hand. Users can see missions, proposals, and live decisions, yet the product still lacks an honest operator view of:

- what is waiting on the user
- what is blocked
- what is scheduled
- what is unresolved
- what should come back onto the agenda next

Wave 4 adds that operator layer without pretending Lango already has calendar, inbox, or external task-system integrations.

## What Changes

This change adds the first work-and-life operating loop slice:

- project loop rows over real existing sources rather than introducing a new durable loop table
- add deterministic agenda ordering for unresolved loops
- keep the first slice honest by using only landed sources
- allow scheduled automation loops from cron jobs only in the first slice
- define explicit deterministic follow-up predicates instead of narrative heuristics
- keep unsupported calendar, inbox, and external task integrations as explicit non-goals

## Scope Guardrails

Wave 4 is intentionally narrow:

- loops are a presentation and coordination model, not a replacement for durable missions
- no fabricated calendar, inbox, or external task-system sources
- no workflow-run scheduled loop source until a dedicated adapter exists
- no broad heuristic prioritization
- no new durable loop database table

## User Impact

Mission Control becomes a more useful operator surface. Users can see a deterministic agenda of real open loops drawn from durable missions, pending inquiries, cron-backed scheduled automation, dead-letter backlog, and explicit follow-up predicates. The slice stays honest about what Lango does not yet integrate.
