## ADDED Requirements

### Requirement: Broker child stderr routing is seam-aware
The storage broker client SHALL route the broker child process stderr through
an explicit writer seam instead of binding command construction directly to
process-global stderr.

#### Scenario: Broker command uses injected stderr writer
- **WHEN** broker command construction receives a stderr writer
- **THEN** the child command SHALL use that writer for `cmd.Stderr`
- **AND** the test SHALL NOT need to start the child process
