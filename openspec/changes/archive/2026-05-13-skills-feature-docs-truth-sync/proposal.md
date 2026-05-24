## Why

The README and architecture docs were already corrected to explain that built-in skills were removed, but `docs/features/skills.md` still claimed that Lango ships 30 embedded default skills. That is the opposite of the current runtime and spec contract, which now uses only a placeholder embedded path unless real skill files are added later.

## What Changes

- Rewrite the skills feature page to describe the current embedded scaffold behavior instead of a built-in default skill bundle.
- Clarify that the default skills directory remains empty until the user creates, imports, or installs skills.
- Extend the project-docs spec so the skills feature page is covered by the same truth requirement as README and project-structure.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `project-docs`: public skill-system documentation now matches the current embedded-skill contract.

## Impact

- Affected docs: `docs/features/skills.md`
- Affected specs: `openspec/specs/project-docs/spec.md`
