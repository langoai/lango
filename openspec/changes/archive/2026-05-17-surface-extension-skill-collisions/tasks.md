## 1. Tests First

- [x] 1.1 Add an app startup regression test for cross-extension skill collision fatal propagation.
- [x] 1.2 Run the focused test and confirm it fails before implementation.

## 2. Implementation

- [x] 2.1 Change `initSkills` to return `(*skill.Registry, error)`.
- [x] 2.2 Propagate `registry.LoadSkills` errors from `intelligenceModule.Init`.
- [x] 2.3 Preserve existing best-effort behavior for default skill deployment warnings and disabled skills.

## 3. Verification

- [x] 3.1 Run focused app and skill tests.
- [x] 3.2 Run `openspec validate surface-extension-skill-collisions --strict`.
- [x] 3.3 Run `go build ./...` and `go test ./...`.
- [x] 3.4 Run subagent review.
- [x] 3.5 Sync and archive the OpenSpec change.
