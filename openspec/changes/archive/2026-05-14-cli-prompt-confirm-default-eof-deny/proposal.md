## Why

The top-level `prompt.Confirm(...)` wrapper already uses the safer EOF-deny path, but the remaining regression name and the main prompt helper spec still describe the older raw-read-error behavior. That leaves an important safety contract under-documented and easy to regress silently.

## What Changes

- Rename and tighten the prompt package regression for default-wrapper EOF denial
- Record the default-wrapper EOF-deny contract in the main prompt helper spec
- Archive the now-accurate delta once validation passes

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-prompt-helpers`: the default confirmation wrapper treats EOF as denial

## Impact

- Affected code: `internal/cli/prompt/prompt_test.go`
- Affected specs: `openspec/specs/cli-prompt-helpers/spec.md`
- No runtime UX change; this makes the existing safer default easier to verify and maintain
