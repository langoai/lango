## 1. Production Context Guard

- [x] 1.1 Add a repository test that rejects `context.TODO()` in production Go files
- [x] 1.2 Replace the remaining test-only `context.TODO()` uses with `context.Background()`
- [x] 1.3 Update OpenSpec coverage for the broader guard

## 2. Verification

- [x] 2.1 Validate with targeted tests, repository-wide build/test, and OpenSpec checks
