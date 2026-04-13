## MODIFIED Requirements

### Requirement: Session timeout annotation
The `session.Store` interface SHALL include `AnnotateTimeout(key string, partial string) error`. On timeout, callers SHALL invoke this to append a synthetic assistant message marking the interrupted turn. Raw partial drafts SHALL NOT be persisted into session history.

#### Scenario: Timeout with no partial response
- **WHEN** a timeout occurs and no partial text was accumulated
- **THEN** AnnotateTimeout SHALL append an assistant message with "[This response was interrupted due to a timeout]"

#### Scenario: Timeout with partial response
- **WHEN** a timeout occurs and partial text was accumulated
- **THEN** AnnotateTimeout SHALL append only the timeout marker
- **AND** it SHALL NOT persist the raw partial text ahead of the marker

#### Scenario: Next turn after timeout
- **WHEN** the user sends a new message after a timeout-annotated turn
- **THEN** the session history SHALL contain a complete user→assistant pair, preventing error leakage
