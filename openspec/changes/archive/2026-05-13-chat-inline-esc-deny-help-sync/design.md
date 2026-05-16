## Context

`handleApprovingKey()` already matches both `d` and `esc` for denial across approval tiers, and the fullscreen dialog plus public docs already reflect that. Only the inline strip still under-describes the same action.

## Goals / Non-Goals

**Goals:**

- Make the inline approval strip advertise the full deny key surface it already supports.

**Non-Goals:**

- Change deny behavior.
- Redesign the inline strip layout.

## Decisions

- Use `[d]/Esc deny` wording in the strip.
  - Rationale: it is compact, mirrors the fullscreen dialog semantics, and exposes the existing alternative key without over-expanding the strip.

## Risks / Trade-offs

- [The strip text becomes slightly wider] → The strip already truncates safely and the added discoverability is worth the small width cost.
