## Context

The embedded settings editor renders `Settings saved` and `Save failed: ...` banners at the top of the menu view after save attempts, clearing them on the next key press. That is already part of the shipped interaction model, but the public feature docs currently omit it.

## Goals / Non-Goals

**Goals:**

- Document where embedded save feedback appears on the Settings page.
- Keep the description short and operator-facing.

**Non-Goals:**

- Change Settings runtime behavior.
- Document every error-path detail beyond the banner location and meaning.

## Decisions

- Add one short sentence to the Settings section rather than a separate subsection.
  - Rationale: this keeps the docs compact while still surfacing the expected save-feedback loop.

## Risks / Trade-offs

- [Docs become slightly more detailed] → The added wording is limited to one visible interaction outcome already present in the product.
