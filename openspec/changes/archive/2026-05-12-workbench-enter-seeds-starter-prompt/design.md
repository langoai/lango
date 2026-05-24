## Overview

This change shortens the workbench first-success path without auto-submitting a model turn.

## Design Decisions

### Enter seeds but does not run

`Enter` on the empty ready-profile workbench loads the first starter prompt into the composer, but still leaves the final model call to the operator. This preserves explicit control while removing a dead keypress.

### No behavior change for incomplete profiles

If the profile still needs setup, `Enter` continues to avoid seeding a starter prompt. The setup-first guidance remains the only path in that state.
