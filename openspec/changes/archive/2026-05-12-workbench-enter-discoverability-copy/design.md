## Overview

This is a copy-level UX alignment change.

## Design Decisions

### Teach the fastest path in the first-screen copy

The empty ready-profile workbench should advertise the lowest-friction path first:

- `Enter` for the default starter prompt
- `1/2/3` for explicit starter selection

The workbench already behaves this way; the UI copy now teaches it directly.

### Keep incomplete-profile guidance unchanged

Setup-first states still prioritize recovery commands and do not mention the `Enter` starter path.
