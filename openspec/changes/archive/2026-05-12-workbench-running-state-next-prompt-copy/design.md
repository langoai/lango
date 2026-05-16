## Overview

This is a copy-level improvement for the in-flight starter state.

## Design Decisions

### Running-state copy should advertise the next best action

When a starter prompt is already running, the operator's next meaningful options are:

- wait for the current result
- cancel
- type the next prompt and press `Enter` to interrupt and redirect

The copy now teaches that third path explicitly.
