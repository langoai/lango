## Why

Cockpit pages are not consistent about how they render vertical navigation hints in the help bar. Tasks, Approvals, and Dead Letters use compact arrow-style labels like `↑/k`, while Sessions and Tools still expose textual `up/k` and `down/j` labels.

## What Changes

- Update Sessions and Tools help bindings to use the same arrow-style labels as the rest of cockpit.
- Add regressions that lock the rendered help labels.
- Update cockpit page specs to require consistent arrow-style vertical navigation hints.

## Impact

- The cockpit help bar reads as one cohesive UI instead of a mix of styles.
- Tests and specs lock the same label contract across pages.
