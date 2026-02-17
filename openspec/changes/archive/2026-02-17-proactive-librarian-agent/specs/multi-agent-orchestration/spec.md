## MODIFIED Requirements

### Requirement: Librarian Agent Specification
The librarian sub-agent SHALL handle knowledge management including: search, RAG, graph traversal, knowledge/skill persistence, and knowledge inquiries. The agent spec SHALL include `librarian_` in its Prefixes list and `inquiry`, `question`, `gap` in its Keywords list. The Instruction SHALL include a "Proactive Behavior" section instructing the agent to weave pending inquiries naturally into responses.

#### Scenario: Librarian tool routing with inquiry prefix
- **WHEN** a tool named `librarian_pending_inquiries` is partitioned
- **THEN** it is assigned to the librarian sub-agent's tool set

#### Scenario: Inquiry keyword routing
- **WHEN** the orchestrator receives a request containing "inquiry" or "gap"
- **THEN** the routing table matches the librarian agent via keyword matching

### Requirement: Capability Map
The orchestrator capability map SHALL include an entry mapping the `librarian_` prefix to "knowledge inquiries and gap detection".

#### Scenario: Capability description includes librarian tools
- **WHEN** capabilityDescription is called for a tool set containing `librarian_pending_inquiries`
- **THEN** the description includes "knowledge inquiries and gap detection"
