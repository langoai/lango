## Overview

This is a copy-level accuracy fix for seeded-starter guidance.

## Design Decisions

### Guidance follows focus state

When a starter is armed:

- if focus is on Composer: say `Enter` submits
- if focus is elsewhere: say `Tab` to Composer, then `Enter`

That keeps the UI language aligned with the actual key routing instead of assuming the operator never moved focus away.
