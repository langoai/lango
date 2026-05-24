## Context

`TestCLIProductionCodeAvoidsRawPrintsAndDirectStdStreams` walks `internal/cli` and rejects raw `fmt.Print*` calls plus direct `os.Stdout`/`os.Stderr` references outside approved seam files. It does not check `os.Stdin`.

## Decision

Extract the scan into a helper that accepts a root directory and allowed relative paths, then add a fixture-style unit test using `t.TempDir()` to verify direct `os.Stdin` is treated as a violation.

## Tradeoffs

The guard still allows approved seam files by path rather than parsing AST usage. That keeps the repository guard simple and aligned with the existing string-based approach.
