## MODIFIED Requirements

### Requirement: Event replay with author mapping
The EventsAdapter SHALL reconstruct events from stored session messages with correct author attribution for the ADK runner.

#### Scenario: Model role mapping
- **WHEN** a stored message has `Role: "model"` and empty `Author`
- **THEN** the EventsAdapter SHALL map the author to `rootAgentName` (or `"lango-agent"` if rootAgentName is empty)
- **AND** the author SHALL NOT be the literal string `"model"`

#### Scenario: Unknown role fallback
- **WHEN** a stored message has an unrecognized `Role` and empty `Author`
- **THEN** the EventsAdapter SHALL map the author to `rootAgentName` (or `"lango-agent"` if rootAgentName is empty)
- **AND** the author SHALL NOT produce "Event from an unknown agent" warnings
