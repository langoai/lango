## 1. Shared Confirmation Integration

- [x] 1.1 Replace the extension-local confirmation parser with the shared prompt helper
- [x] 1.2 Preserve the existing non-TTY refusal behavior for scripted runs without `--yes`

## 2. Verification

- [x] 2.1 Add command-level extension regressions for confirm, deny, and non-TTY warning paths
- [x] 2.2 Update the public extension-pack documentation note in `README.md`
- [x] 2.3 Validate with package tests, repository-wide build/test, and OpenSpec checks
