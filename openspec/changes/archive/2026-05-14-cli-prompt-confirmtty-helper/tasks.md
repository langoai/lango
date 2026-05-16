## 1. Shared Guarded Confirmation

- [x] 1.1 Add a terminal-guarded confirmation helper to the prompt package
- [x] 1.2 Replace the extension-local confirmation wrapper with the shared helper

## 2. Verification

- [x] 2.1 Add prompt-package tests for non-TTY rejection and EOF-as-deny behavior
- [x] 2.2 Update extension regression coverage to rely on the shared helper contract
- [x] 2.3 Validate with package tests, repository-wide build/test, and OpenSpec checks
