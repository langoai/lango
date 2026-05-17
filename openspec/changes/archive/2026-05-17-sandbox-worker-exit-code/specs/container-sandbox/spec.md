## MODIFIED Requirements

### Requirement: Sandbox Docker image
A `build/sandbox/Dockerfile` MUST define a minimal Debian-based image with the lango binary, running as non-root `sandbox` user with `--sandbox-worker` entrypoint. The sandbox worker entrypoint SHALL expose testable exit-code-returning execution while preserving process exit-code semantics when launched as a binary.

#### Scenario: Sandbox worker preserves protocol exit codes
- **WHEN** the sandbox worker receives malformed input or an unregistered tool
- **THEN** worker execution SHALL return exit code `1` and write a JSON error result

#### Scenario: Sandbox worker tool errors stay JSON-level errors
- **WHEN** a registered worker tool returns an error
- **THEN** worker execution SHALL return exit code `0` and write a JSON error result

#### Scenario: Sandbox worker success writes JSON output
- **WHEN** a registered worker tool succeeds
- **THEN** worker execution SHALL return exit code `0` and write a JSON output result
