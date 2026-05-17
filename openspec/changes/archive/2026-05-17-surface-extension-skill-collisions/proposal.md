## Why

Extension-authored skills can reach an invalid runtime state when two installed packs provide the same skill name under separate `ext-<pack>/` directories. The file skill store already detects this and returns an error, but app startup currently logs the load failure as a warning and continues with a partially initialized skill registry.

That behavior violates the existing `skill-system` requirement that cross-extension skill collisions are not resolvable at runtime and must be fatal. Continuing after a collision hides a broken extension install and can leave operators with missing or unpredictable skill tools.

## What Changes

- Propagate skill registry load errors from app intelligence module startup instead of warning and continuing.
- Keep non-fatal default skill deployment warnings unchanged.
- Add an app-level regression test that builds a real extension registry and colliding extension skill directories, then verifies startup returns a fatal error naming the collision.

## Impact

- Startup fails fast when installed extension packs collide on skill names.
- Existing valid extension installs are unaffected.
- Public docs do not need new user-facing command documentation; this aligns runtime behavior with the already documented/spec'd extension collision contract.
