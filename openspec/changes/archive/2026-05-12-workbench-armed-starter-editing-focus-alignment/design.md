## Overview

This is a small interaction-consistency fix for the seeded-starter state.

## Design Decisions

### Editing keys imply composer intent

If the operator presses a composer editing key while a starter prompt is armed, that is strong evidence they intend to edit the prompt rather than stay on the current lane. The workbench now treats that as an implicit return to `Composer`.
