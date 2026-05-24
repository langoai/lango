## ADDED Requirements

### Requirement: Cmd entrypoint stream routing stays disciplined
Top-level binary entrypoints SHALL avoid raw print calls and direct standard-stream references except where explicit seam declarations intentionally define default process streams.

#### Scenario: Cmd entrypoints reject raw print calls
- **WHEN** a non-test Go file under `cmd/` reintroduces `fmt.Print`, `fmt.Printf`, or `fmt.Println`
- **THEN** an executable repository test SHALL fail

#### Scenario: Cmd entrypoints reject direct standard streams outside seams
- **WHEN** a non-test Go file under `cmd/` reintroduces direct `os.Stdin`, `os.Stdout`, or `os.Stderr` references outside the approved seam declaration lines
- **THEN** an executable repository test SHALL fail
