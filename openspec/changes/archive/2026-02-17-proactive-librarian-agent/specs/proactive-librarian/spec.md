## ADDED Requirements

### Requirement: Inquiry Ent Schema
The system SHALL persist knowledge inquiries using an Ent schema with fields: id (UUID), session_key, topic, question, context (optional), priority (low/medium/high), status (pending/resolved/dismissed), answer (optional), knowledge_key (optional), source_observation_id (optional), created_at, resolved_at (optional). The schema SHALL have indexes on (session_key, status) and (status).

#### Scenario: Inquiry creation
- **WHEN** the proactive buffer detects a knowledge gap
- **THEN** a new Inquiry record is created with status "pending" and the gap's topic, question, context, and priority

#### Scenario: Inquiry resolution
- **WHEN** a user's answer is matched to a pending inquiry
- **THEN** the inquiry status is set to "resolved", the answer is stored, and resolved_at is set

### Requirement: LibrarianConfig
The system SHALL provide a `LibrarianConfig` struct with fields: Enabled (bool), ObservationThreshold (int, default 2), InquiryCooldownTurns (int, default 3), MaxPendingInquiries (int, default 2), AutoSaveConfidence (string, default "high"), Provider (string), Model (string). The config SHALL be accessible at `config.Librarian`.

#### Scenario: Librarian disabled
- **WHEN** `librarian.enabled` is false
- **THEN** no proactive librarian components are initialized

#### Scenario: Default config values
- **WHEN** config values are zero/empty
- **THEN** defaults are applied: ObservationThreshold=2, InquiryCooldownTurns=3, MaxPendingInquiries=2, AutoSaveConfidence="high"

### Requirement: Observation Analyzer
The system SHALL analyze conversation observations via LLM to extract knowledge (with type, category, content, confidence, key) and detect knowledge gaps (with topic, question, context, priority). The analyzer SHALL output a structured AnalysisOutput containing extractions and gaps arrays.

#### Scenario: Successful analysis
- **WHEN** observations are passed to the analyzer
- **THEN** the LLM returns JSON with extractions and gaps, parsed into AnalysisOutput

#### Scenario: Empty observations
- **WHEN** zero observations are provided
- **THEN** an empty AnalysisOutput is returned without LLM call

### Requirement: Inquiry Processor
The system SHALL detect user answers to pending inquiries by analyzing recent messages via LLM. When a match is detected with high or medium confidence, the system SHALL save the answer as structured knowledge and resolve the inquiry.

#### Scenario: Answer detected
- **WHEN** a recent message matches a pending inquiry with high confidence
- **THEN** knowledge is saved via knowledge.Store.SaveKnowledge and the inquiry is resolved

#### Scenario: No pending inquiries
- **WHEN** no pending inquiries exist for the session
- **THEN** the processor returns immediately without LLM call

### Requirement: Proactive Buffer
The system SHALL provide an async ProactiveBuffer with Start/Trigger/Stop lifecycle. On each trigger, the buffer SHALL: (1) process pending inquiry answers, (2) analyze observations if threshold is met, (3) auto-save high-confidence extractions, (4) create inquiries from gaps respecting cooldown and max-pending limits.

#### Scenario: Turn complete triggers buffer
- **WHEN** gateway.OnTurnComplete fires for a session
- **THEN** the ProactiveBuffer.Trigger is called with the session key

#### Scenario: Cooldown prevents inquiry creation
- **WHEN** turns since last inquiry is less than InquiryCooldownTurns
- **THEN** no new inquiries are created even if gaps are detected

#### Scenario: Max pending limit
- **WHEN** pending inquiry count reaches MaxPendingInquiries
- **THEN** no additional inquiries are created

### Requirement: Inquiry Store CRUD
The system SHALL provide InquiryStore with methods: SaveInquiry, ListPendingInquiries, ResolveInquiry, DismissInquiry, CountPendingBySession. ListPendingInquiries SHALL filter by session_key and status=pending, ordered by created_at.

#### Scenario: List pending inquiries
- **WHEN** ListPendingInquiries is called with a session key and limit
- **THEN** only pending inquiries for that session are returned, up to the limit

#### Scenario: Dismiss inquiry
- **WHEN** DismissInquiry is called with an inquiry ID
- **THEN** the inquiry status is set to "dismissed" and resolved_at is set

### Requirement: Librarian Agent Tools
The system SHALL expose two new agent tools: `librarian_pending_inquiries` (list pending inquiries for a session) and `librarian_dismiss_inquiry` (dismiss a pending inquiry by ID). Both tools SHALL be partitioned to the librarian sub-agent via the `librarian_` prefix.

#### Scenario: List pending inquiries tool
- **WHEN** agent calls librarian_pending_inquiries with a session key
- **THEN** pending inquiries are returned with count

#### Scenario: Dismiss inquiry tool
- **WHEN** agent calls librarian_dismiss_inquiry with an inquiry UUID
- **THEN** the inquiry is dismissed and confirmation is returned

### Requirement: Auto-save Knowledge from Extractions
The system SHALL automatically save knowledge extractions that meet the configured auto-save confidence threshold. High-confidence extractions are saved without user confirmation. Optional graph triples (subject/predicate/object) SHALL be forwarded via the graph callback if available.

#### Scenario: High confidence auto-save
- **WHEN** an extraction has confidence "high" and AutoSaveConfidence is "high"
- **THEN** the extraction is saved as knowledge via knowledge.Store.SaveKnowledge

#### Scenario: Below threshold extraction
- **WHEN** an extraction has confidence "medium" and AutoSaveConfidence is "high"
- **THEN** the extraction is NOT auto-saved
