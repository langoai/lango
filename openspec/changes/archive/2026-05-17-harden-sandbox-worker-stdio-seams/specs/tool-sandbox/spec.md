## ADDED Requirements

### Requirement: Sandbox worker public wrapper stdio is seam-aware

The sandbox worker public wrapper SHALL preserve production stdin/stdout defaults while routing through testable package-level stdio seams.

#### Scenario: Public worker wrapper uses injected standard streams

- **WHEN** `RunWorker` is invoked with worker stdin and stdout seams replaced by in-memory streams
- **THEN** it SHALL decode the JSON `ExecutionRequest` from the injected stdin seam
- **AND** it SHALL write exactly one JSON `ExecutionResult` to the injected stdout seam

#### Scenario: Explicit worker IO protocol remains unchanged

- **WHEN** callers invoke `RunWorkerWithIO` with explicit input and output streams
- **THEN** request decoding, result encoding, and exit-code behavior SHALL continue to use those explicit streams without depending on the public wrapper seams
