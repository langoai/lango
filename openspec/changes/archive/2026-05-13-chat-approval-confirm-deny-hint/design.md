## Context

`handleApprovingKey()` already resets confirm-pending and falls through to denial when `d` or `Esc` is pressed, so the underlying control flow is correct. The missing piece is visible guidance during confirm-pending states.

## Goals / Non-Goals

**Goals:**

- Preserve confirm-key guidance while also surfacing the deny escape path.

**Non-Goals:**

- Change approval behavior.
- Introduce new approval keys.

## Decisions

- Append a short `d/Esc denies` hint to both confirm-pending prompts.
  - Rationale: this keeps the prompt compact while making the live denial path explicit.

## Risks / Trade-offs

- [Confirm prompt becomes a bit longer] → The extra words are justified because they surface a valid high-safety exit path.
