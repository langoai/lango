## Why

The cockpit feature reference has increasingly detailed sections for several pages, but Mission Control still does not explicitly enumerate its current first-screen key surface. The runtime already has a stable help contract around `tab`, `↑/↓`, `enter`, proposal actions, and the empty-state reduction, so the public feature docs should expose that operator surface directly.

## What Changes

- Add a Mission Control key-surface subsection to `docs/features/cockpit.md`.
- Describe the current help-bar behavior, including empty-state help reduction and proposal-action keys.
- Extend downstream docs-sync requirements so that Mission Control key guidance stays aligned.

## Impact

- Public cockpit docs describe the current Mission Control operator surface more concretely.
- Future help-surface drift becomes easier to detect in docs reviews.
