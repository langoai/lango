## ADDED Requirements

### Requirement: CLI production code uses command-stream-safe output paths
CLI production code SHALL avoid raw process-global print calls and direct standard-stream references except where explicit seam defaults are intentionally defined.

#### Scenario: CLI production code rejects raw fmt.Print calls
- **WHEN** a non-test Go file under `internal/cli` reintroduces `fmt.Print`, `fmt.Printf`, or `fmt.Println`
- **THEN** an executable repository test SHALL fail

#### Scenario: CLI production code rejects direct standard streams outside seams
- **WHEN** a non-test Go file under `internal/cli` reintroduces direct `os.Stdout` or `os.Stderr` references outside approved seam files
- **THEN** an executable repository test SHALL fail
