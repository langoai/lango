## Overview

This is a small interaction refinement for the seeded-starter state.

## Design Decisions

### Numeric starter selection stays live after seeding

The numeric starter shortcuts are now valid in both seeded and unseeded empty-workbench states. That keeps the starter selection model consistent:

- first use: choose a starter
- later use: replace the current starter

instead of changing semantics once the first prompt is armed.
