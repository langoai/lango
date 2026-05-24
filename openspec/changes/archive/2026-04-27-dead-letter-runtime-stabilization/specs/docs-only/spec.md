## MODIFIED Requirements

### Requirement: Dead-letter browsing / status observation page describes the landed operator surface
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the dead-letter CLI and cockpit operator surfaces as they actually behave.

#### Scenario: Dead-letter docs describe stabilized retry and cockpit wiring
- **WHEN** a user reads `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find that CLI and cockpit retry inject a local default operator principal when the runtime context is otherwise empty
- **AND** they SHALL find that cockpit dead-letter filter fields are forwarded through the shell adapter into the dead-letter list tool
- **AND** they SHALL find that dispatch-family grouping uses the shared classifier across CLI and cockpit

### Requirement: P2P knowledge exchange track reflects landed dead-letter/runtime stabilization
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe dead-letter browsing / status observation as landed work with the stabilized cockpit shell adapter forwarding, retry principal injection, and shared dispatch-family grouping.

#### Scenario: Track page includes stabilized dead-letter/runtime behavior
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find cockpit filter forwarding described as landed behavior
- **AND** they SHALL find retry principal injection described as landed behavior
- **AND** they SHALL find shared dispatch-family grouping between CLI and cockpit described as landed behavior
