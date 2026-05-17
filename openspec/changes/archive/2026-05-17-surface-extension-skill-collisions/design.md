## Context

`internal/skill.FileSkillStore.ListActive` already returns an error when multiple allowed extension packs provide the same skill name. `internal/app.initSkills` currently catches `registry.LoadSkills` errors, logs `load skills error`, and still returns the registry.

The caller that represents startup wiring is `intelligenceModule.Init`, which has an error return and can fail module initialization.

## Decision

Change `initSkills` from returning only `*skill.Registry` to returning `(*skill.Registry, error)`.

- If skills are disabled, return `(nil, nil)`.
- If default skill deployment fails, keep the current warning-only behavior because that is not the collision path and existing behavior is intentionally best-effort.
- If `registry.LoadSkills` fails, return a wrapped error such as `load skills: ...`.
- In `intelligenceModule.Init`, propagate that error so app startup fails.

## Testing

Add an app-level test that:

- Creates two valid installed extension packs in a temp extensions directory.
- Creates two corresponding `ext-<pack>/` skill directories under a temp skills directory with the same `SKILL.md` name.
- Loads the extension registry so both packs are allowed.
- Runs `intelligenceModule.Init`.
- Asserts the returned error includes the colliding skill name and both extension packs.
