## ADDED Requirements

### Requirement: Core production stream guards stay executable

Repository-level regressions that introduce unapproved direct process-global standard-stream references in non-CLI core production packages SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Core direct stdio regressions are rejected

- **WHEN** a production Go file under `internal/` outside `internal/cli/`, `internal/testutil/`, and generated code directly references `os.Stdin`, `os.Stdout`, or `os.Stderr`
- **THEN** an executable test SHALL fail unless that exact file and line fragment is explicitly allowlisted as an intentional seam

#### Scenario: Existing stream guard ownership remains separated

- **WHEN** stream guards scan the repository
- **THEN** `cmd/` entrypoints SHALL remain owned by the cmd entrypoint stream guard
- **AND** `internal/cli/` production code SHALL remain owned by the CLI production stream guard
