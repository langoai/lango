# Tasks

## 1. Planning

- [x] 1.1 Identify stale logging output path copy after the stderr default change.
- [x] 1.2 Limit scope to settings UI text, public config docs, and executable coverage.

## 2. Tests

- [x] 2.1 Add a failing settings form test for the stderr fallback copy.

## 3. Implementation

- [x] 3.1 Update the Logging settings form placeholder and description.
- [x] 3.2 Update README and configuration docs for `logging.outputPath`.

## 4. Verification

- [x] 4.1 Run focused settings form tests.
- [x] 4.2 Run `go build ./...`.
- [x] 4.3 Run `go test ./...`.
- [x] 4.4 Run `openspec validate --all --strict`.
- [x] 4.5 Archive the OpenSpec change after implementation is verified.
