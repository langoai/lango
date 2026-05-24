## ADDED Requirements

### Requirement: Architecture landing and track docs reference background post-adjudication execution
The architecture landing page and P2P knowledge-exchange track doc SHALL reference the landed background post-adjudication execution slice.

#### Scenario: Landing page links background post-adjudication execution
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** they SHALL see the Background Post-Adjudication Execution page listed with the other architecture pages

#### Scenario: Track doc reflects landed background post-adjudication execution
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** background post-adjudication execution SHALL be described as landed slice work with retry orchestration, dead-letter handling, dedicated status observation, and policy-driven defaults still remaining
