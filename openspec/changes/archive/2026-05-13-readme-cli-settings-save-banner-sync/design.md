## Context

The runtime and cockpit feature docs already treat inline save feedback as part of the embedded Settings experience. The README shortcut table and CLI overview are still lagging artifacts.

## Goals / Non-Goals

**Goals:**

- Surface Settings save-feedback behavior in the README and CLI overview.
- Keep the wording concise and overview-friendly.

**Non-Goals:**

- Change Settings runtime behavior.
- Duplicate the full feature-doc section.

## Decisions

- Add the save-feedback clause directly to the existing Settings descriptions.
  - Rationale: the overview artifacts already have a natural Settings slot; they just need the missing detail.

## Risks / Trade-offs

- [Overview lines become slightly longer] → The added phrasing is small and directly useful to operators.
