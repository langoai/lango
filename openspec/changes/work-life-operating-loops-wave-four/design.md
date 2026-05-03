## Context

Wave 3 made Mission Control proactive by introducing transient proposals, but the surface is still centered on individual missions and point-in-time decisions. The approved Wave 4 direction adds a higher-order operating view: recurring work, unresolved inquiries, dead-letter backlog, scheduled automation, and deterministic follow-up signals should all be visible as operator loops.

The first slice is intentionally narrow. It must project loops only from sources that already exist in the codebase today, remain deterministic in terminal form, and avoid creating a second durable system that competes with `Mission`.

## Goals / Non-Goals

**Goals:**
- add a projection-first `LoopView` / agenda surface on top of existing sources
- make loop ordering deterministic and testable
- keep durable missions visible while adding loop coordination context
- limit the first scheduled automation slice to cron-job sources only
- generate follow-up loops only from explicit deterministic predicates over existing mission/proposal/session facts

**Non-Goals:**
- adding a durable loop database table
- replacing durable missions with loops
- introducing broad heuristic prioritization
- implying calendar, inbox, or external task-system integrations that do not exist
- enabling workflow-run loop sources before a dedicated adapter exists

## Decisions

### 1. Projection-first loops, no new durable loop table

Wave 4 stays projection-first. The first slice should derive `LoopView` rows from existing durable and transient sources rather than adding another persistent table. This keeps the change narrow and avoids competing truths with the existing `mission` layer.

Alternative considered:
- adding a durable `loop` table now
Why rejected:
- it would create a second long-lived coordination model before loop semantics are stable

### 2. Real-source-only first slice

Loop rows must come only from sources that already exist in code:
- durable missions
- pending knowledge inquiries
- dead-letter / retry backlog
- cron-job schedule state
- deterministic follow-up signals from recent mission/proposal/session facts

Unsupported sources must remain absent rather than fabricated.

Alternative considered:
- including placeholder calendar, inbox, or external task sources
Why rejected:
- it would make the operator surface dishonest and expand scope without real adapters

### 3. Deterministic agenda ordering

Agenda ordering will follow the approved Wave 4 order:
1. waiting-user
2. blocked
3. active
4. scheduled
5. needs-review
6. resolved

Within the same category, newer updates sort first. The first slice does not add broader heuristic ranking.

Alternative considered:
- heuristic “importance” scoring
Why rejected:
- it would be difficult to explain and impossible to validate reliably in terminal form

### 4. Scheduled automation is cron-only in the first slice

Although the higher-level design leaves room for broader automation loops later, this first implementation only projects scheduled automation from cron-job sources. Workflow-run loops remain deferred until their own adapter exists.

Alternative considered:
- including workflow runs immediately
Why rejected:
- the current slice requires a dedicated adapter to avoid fabricating unstable schedule semantics

### 5. Follow-up loops use explicit predicates only

Follow-up loops are derived only from deterministic facts:
- accepted proposal with no active linked execution yet
- completed mission updated recently and still needing review
- failed or blocked recurring cron automation
- unresolved inquiry older than a threshold

Alternative considered:
- narrative or heuristic follow-up generation
Why rejected:
- it would blur the surface into a vague productivity dashboard

## Risks / Trade-offs

- [Risk] Loop projection duplicates some mission/proposal visibility.
  → Mitigation: keep durable missions as the primary owned work surface and make loops an additive coordination layer.

- [Risk] Source availability differs across deployments.
  → Mitigation: unsupported or unavailable sources remain absent; no fake placeholder integrations.

- [Risk] Agenda ordering could become opaque if it grows beyond deterministic rules.
  → Mitigation: keep first-slice ordering fixed and source-native.

- [Risk] Cron-only scheduling may feel incomplete.
  → Mitigation: document the explicit limitation and defer workflow-run scheduling until a real adapter exists.

## Migration Plan

1. Add loop projection types and source adapters without changing durable mission truth.
2. Wire loop readers into the app boundary only when real sources exist.
3. Extend Mission Control with a compact loop/agenda surface while leaving the existing mission board intact.
4. Update docs to describe the real loop source set and explicit non-goals.

Rollback is low-risk because the slice is additive and projection-first. Disabling the loop surface returns Mission Control to its mission/proposal/decision view without data migration.

## Open Questions

- Whether the first loop lane lands as a compact band, toggle, or secondary pane should follow the implementation slice but stay within the “no major layout rewrite” constraint.
- The exact “needs review” recent-threshold can be finalized in implementation as long as it remains deterministic and source-native.
