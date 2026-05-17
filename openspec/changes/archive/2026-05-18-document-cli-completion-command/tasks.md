# Tasks

## 1. Planning

- [x] 1.1 Audit root CLI help and current public docs for `lango completion` coverage.
- [x] 1.2 Add focused OpenSpec artifacts for completion command documentation coverage.
- [x] 1.3 Validate the OpenSpec change before implementation.
- [x] 1.4 Commit the OpenSpec planning artifacts separately.

## 2. Regression Test

- [x] 2.1 Extend README top-level utility docs guard to require `lango completion`.
- [x] 2.2 Add a CLI reference docs guard requiring `lango completion`.
- [x] 2.3 Confirm the focused docs guard fails against the current docs.

## 3. Documentation

- [x] 3.1 Document `lango completion` in `README.md`.
- [x] 3.2 Document `lango completion` in `docs/cli/index.md`.
- [x] 3.3 Avoid documenting behavior beyond Cobra's shipped completion command.

## 4. Review

- [x] 4.1 Request teammate review for docs accuracy and guard quality.
- [x] 4.2 Address actionable findings before archiving.

## 5. Verification

- [x] 5.1 Run focused `internal/testutil` docs guard tests.
- [x] 5.2 Run `go build ./...`.
- [x] 5.3 Run `go test ./...`.
- [x] 5.4 Run `openspec validate --all --strict`.
- [x] 5.5 Run `git diff --check`.
- [x] 5.6 Archive the change after verification.
