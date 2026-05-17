## ADDED Requirements

### Requirement: Shared interactive guard supports explicit input streams
The shared CLI prompt package SHALL provide an interactive guard helper that
validates a caller-supplied input stream and returns a caller-supplied guidance
error when that stream is a non-interactive terminal file.

#### Scenario: Explicit guard rejects non-terminal file input
- **WHEN** the explicit interactive guard receives a non-terminal `*os.File`
  input stream
- **THEN** it SHALL return the caller-supplied guidance error

#### Scenario: Explicit guard allows injected readers
- **WHEN** the explicit interactive guard receives an injected non-file reader
  such as `bytes.Buffer`
- **THEN** it SHALL return nil so command tests and embedded wrappers can drive
  the subsequent interaction through explicit streams
