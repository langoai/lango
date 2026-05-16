## 1. Prompt Helper Seams

- [x] 1.1 Add injectable default input/output seams to `prompt.Confirm(...)`
- [x] 1.2 Keep `ConfirmIO(...)` as the shared confirmation implementation
- [x] 1.3 Restore shared config/bootstrap loader helpers needed by existing CLI regression tests

## 2. Verification

- [x] 2.1 Add prompt package regressions for wrapper approval, denial, and read-error paths
- [x] 2.2 Validate the change with prompt package tests and strict OpenSpec validation
- [x] 2.3 Re-run repository-wide build and test verification after helper restoration
