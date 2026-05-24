## MODIFIED Requirements

### Requirement: Emergency compaction trigger measurement
The emergency compaction trigger SHALL measure the TOTAL context size including conversation history (`req.Contents`), base prompt tokens, and all injected sections (Knowledge, RAG, Memory, RunSummary). The threshold comparison SHALL use `modelWindow × 0.9`.

#### Scenario: Long conversation triggers compaction
- **WHEN** a session has 100K tokens of conversation history, 8K base prompt, and 5K injected context
- **AND** the model window is 128K tokens
- **THEN** `totalMeasured` SHALL be approximately 113K (100K + 8K + 5K)
- **AND** compaction SHALL trigger because 113K > 115.2K × 0.9

#### Scenario: Short conversation with heavy context does not over-trigger
- **WHEN** a session has 5K tokens of conversation history, 8K base prompt, and 20K injected context
- **AND** the model window is 128K tokens
- **THEN** `totalMeasured` SHALL be approximately 33K
- **AND** compaction SHALL NOT trigger because 33K < 115.2K
