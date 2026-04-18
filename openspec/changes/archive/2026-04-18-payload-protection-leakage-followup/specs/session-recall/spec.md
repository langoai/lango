## ADDED Requirements

### Requirement: Session recall stores only redacted summaries
Session recall indexing MUST use decrypted message content as summarization input but MUST store only redacted summary text in the recall table.

#### Scenario: Recall pipeline summarizes decrypted content
- **WHEN** a session with protected messages is ended and recall indexing runs
- **THEN** the summarization step reads decrypted original content
- **AND** the recall table stores only the redacted summary output

#### Scenario: Tool-call plaintext does not leak into recall rows
- **WHEN** a protected message contains tool-call input, output, or thought text
- **THEN** those original values are not written directly to the recall FTS row

#### Scenario: Decrypt failure does not promote projection text
- **WHEN** a protected message row cannot be decrypted during recall generation
- **THEN** the pipeline does not treat the stored plaintext projection as the original message body
