## Context

The runtime ownership migration must not jump directly to all mutators. The lowest-risk path is to move read-only surfaces first, validate that production app and CLI can operate through broker-backed readers, and only then move write paths and store replacement.

## Goals / Non-Goals

**Goals:**
- Add broker RPCs for runtime readers used by app/CLI.
- Replace production reader-side raw handle usage with broker-backed capabilities.
- Keep user-facing CLI output stable.

**Non-Goals:**
- Removing every remaining raw-handle escape hatch in this first slice.
- Moving knowledge/learning/inquiry/agent memory mutators yet.
- Finishing payment factory decomposition.

## Decisions

- This first runtime slice is reader-only.
- Reader order is fixed:
  1. broker reader RPC
  2. storage facade reader
  3. production app/CLI reader switch
  4. retire reader-side raw accessor usage
- Any mutator or service-construction refactor that still needs storage ownership is deferred to the next slice.

## Risks / Trade-offs

- Some transitional raw-handle methods may still exist after this slice, but production reader paths should no longer depend on them.
- Broker reader coverage increases protocol surface area, but it is the necessary step before full runtime ownership.
