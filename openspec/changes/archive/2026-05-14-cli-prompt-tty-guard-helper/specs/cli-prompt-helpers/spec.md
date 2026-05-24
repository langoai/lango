## ADDED Requirements

### Requirement: Shared TTY-input guard supports command-specific guidance
The shared CLI prompt package SHALL provide a helper that rejects non-terminal `*os.File` input streams with a caller-supplied guidance error while allowing other reader types to proceed for tests and explicit injected-input flows.

#### Scenario: TTY-input guard rejects non-terminal file input
- **WHEN** the shared TTY-input guard receives a non-terminal `*os.File` input stream
- **THEN** it SHALL return the caller-supplied guidance error

#### Scenario: TTY-input guard allows injected non-file readers
- **WHEN** the shared TTY-input guard receives an injected non-file reader such as `bytes.Buffer`
- **THEN** it SHALL return nil so command tests can continue driving the interaction through explicit streams
