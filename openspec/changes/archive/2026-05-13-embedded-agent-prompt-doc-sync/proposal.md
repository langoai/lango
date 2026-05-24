## Why

The runtime now embeds built-in per-agent `prompts/agents/<name>/IDENTITY.md` files and uses them as the preferred source for built-in teammate instructions. But the prompt embed test and embedded-prompt-files docs/spec still described only the four top-level prompt files.

## What Changes

- Extend the prompt embed test to assert that built-in per-agent `IDENTITY.md` files are embedded.
- Update the system-prompts feature doc to mention embedded per-agent identities.
- Sync the embedded-prompt-files spec with the expanded embed surface.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `embedded-prompt-files`: docs/spec and tests now match the expanded embedded prompt filesystem surface.

## Impact

- Affected tests: `prompts/embed_test.go`
- Affected docs: `docs/features/system-prompts.md`
- Affected specs: `openspec/specs/embedded-prompt-files/spec.md`
