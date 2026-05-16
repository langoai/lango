## 1. Implementation

- [x] 1.1 Add an internal non-interactive acquisition helper with injected stderr
- [x] 1.2 Keep `AcquireNonInteractive(...)` behavior unchanged via delegation

## 2. Tests

- [x] 2.1 Add warning-path coverage for non-`ErrNotFound` keyring failures
- [x] 2.2 Add silent-fallback coverage for `ErrNotFound`

## 3. Downstream

- [x] 3.1 Update OpenSpec passphrase-acquisition spec with the non-interactive seam contract
- [x] 3.2 Add change proposal
- [x] 3.3 Add change design
- [x] 3.4 Add delta spec
