## Overview

This is a key-routing simplification for the seeded-starter state.

## Design Decisions

### Armed starter prompt becomes a global empty-workbench Enter action

When all of these are true:

- the surface is the standalone workbench
- the workbench is otherwise empty
- a starter prompt is armed in the composer

`Enter` submits the starter prompt regardless of which lane currently has focus.

### Copy follows behavior

Because focus no longer blocks submission, seeded-state guidance no longer tells the operator to tab back to `Composer` before pressing `Enter`.
