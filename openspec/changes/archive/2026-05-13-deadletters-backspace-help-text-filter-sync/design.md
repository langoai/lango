## Context

Dead Letters exposes several text filter fields and routes `Backspace` through `backspaceActiveField()`, which already edits the currently focused text field. The mismatch is not in the runtime path but in the help label, which still implies the key only edits the query field.

## Goals / Non-Goals

**Goals:**

- Make the Dead Letters help bar describe the real `Backspace` behavior.
- Lock the wording with a focused regression.

**Non-Goals:**

- Change Dead Letters filter semantics.
- Add new keys or alter field focus behavior.

## Decisions

- Update the `Backspace` help description from `query` to `text filter`.
  - Rationale: this matches both the existing runtime behavior and the public docs without over-explaining the active-field mechanism in the compressed help bar.

- Add a direct ShortHelp assertion.
  - Rationale: this is the narrowest test for the operator-facing contract.

## Risks / Trade-offs

- [Help wording becomes slightly broader than the original one-word hint] → Use concise `text filter` phrasing so the help bar stays compact while remaining truthful.
