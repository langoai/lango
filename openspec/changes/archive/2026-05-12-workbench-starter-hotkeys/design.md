## Overview

This change turns starter prompts from passive copy into actionable quick-start controls for the standalone workbench.

## Design Decisions

### Hotkeys load prompts instead of auto-submitting them

Pressing `1`, `2`, or `3` loads the corresponding starter prompt into the composer, but does not immediately run a turn. This keeps the quick-start path shorter without surprising the operator with an immediate model call.

### Gated to the empty ready-profile workbench

The hotkeys only apply when all of these are true:

- the surface is the standalone workbench
- the profile is ready
- Mission Control is empty
- the composer is still empty

Outside that state, numeric keys continue behaving like normal text input.

### Copy stays explicit

The body copy, empty composer placeholder, and footer hint all state that `1`, `2`, and `3` are the starter-prompt path so the affordance is visible before the user experiments.
