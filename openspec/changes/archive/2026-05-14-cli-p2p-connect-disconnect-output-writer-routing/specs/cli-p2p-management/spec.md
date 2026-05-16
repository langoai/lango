## ADDED Requirements

### Requirement: P2P connect and disconnect output routing

`lango p2p connect` and `lango p2p disconnect` SHALL write success confirmations through the Cobra command output stream so wrappers and test harnesses can capture them without intercepting process-global stdout.

#### Scenario: Connect confirmation writes to command output
- **WHEN** user runs `lango p2p connect /ip4/1.2.3.4/tcp/9000/p2p/QmPeerId`
- **THEN** the command writes `Connected to peer QmPeerId` to the Cobra command output stream

#### Scenario: Disconnect confirmation writes to command output
- **WHEN** user runs `lango p2p disconnect QmPeerId`
- **THEN** the command writes `Disconnected from peer QmPeerId` to the Cobra command output stream
