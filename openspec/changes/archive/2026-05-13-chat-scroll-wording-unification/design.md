## Context

`PgUp/PgDn` already perform transcript scrolling, but the help text is named inconsistently across the in-product `/help` output and the public cockpit docs. This is purely a wording alignment issue.

## Goals / Non-Goals

**Goals:**

- Use one operator-facing phrase for transcript scrolling.

**Non-Goals:**

- Change scrolling behavior.
- Broaden the change beyond the `PgUp/PgDn` wording.

## Decisions

- Standardize on `Scroll transcript`.
  - Rationale: it is shorter than `Scroll the transcript viewport` but still precise enough to distinguish it from tool/diff scrolling.

## Risks / Trade-offs

- [Slight wording churn in docs/help] → The benefit is a cleaner, more consistent operator surface.
