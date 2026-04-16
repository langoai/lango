## ADDED Requirements

### Requirement: Runtime wake boundary definition
The system SHALL document the boundary between application-layer resume (opt-in `confirmResume + resumeRunId` handshake) and runtime-layer wake (harness re-initialization from persisted state). The design document SHALL enumerate the state categories that must persist for wake to be possible without a full bootstrap pipeline.

#### Scenario: State categories enumerated
- **WHEN** the design document is reviewed
- **THEN** it explicitly covers: in-flight tool call state, pending approval state, supervisor/ADK session bridge state, and crypto provider re-initialization
- **AND** it maps which of these the current resume protocol covers vs does not cover

#### Scenario: No runtime behavior change
- **WHEN** this change is implemented
- **THEN** no runtime code paths for session handling, resume, or bootstrap are modified
- **AND** the change is limited to design documentation and diagnostic tooling
