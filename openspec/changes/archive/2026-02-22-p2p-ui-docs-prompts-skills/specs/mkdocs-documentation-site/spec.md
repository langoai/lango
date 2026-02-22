## MODIFIED Requirements

### Requirement: Navigation includes P2P pages
The mkdocs.yml navigation SHALL include "P2P Network: features/p2p-network.md" in the Features section and "P2P Commands: cli/p2p.md" in the CLI Reference section.

#### Scenario: P2P feature in nav
- **WHEN** the mkdocs site is built
- **THEN** the Features navigation section includes a "P2P Network" entry after "A2A Protocol"

#### Scenario: P2P CLI in nav
- **WHEN** the mkdocs site is built
- **THEN** the CLI Reference navigation section includes a "P2P Commands" entry after "Payment Commands"
