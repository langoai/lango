# Conversation Analysis Specification

## Purpose

Capability spec for conversation-analysis. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Conversation knowledge extraction
The ConversationAnalyzer and SessionLearner SHALL extract knowledge using all 6 categories (rule, definition, preference, fact, pattern, correction) with temporal classification. ALL extracted types SHALL be saved as Knowledge entries. Pattern and correction types SHALL additionally be saved as Learning entries (dual-save).

#### Scenario: All 6 types saved as knowledge
- **WHEN** the ConversationAnalyzer extracts a result with type "rule"
- **THEN** the result SHALL be saved as a Knowledge entry with category "rule"

#### Scenario: Pattern dual-saved
- **WHEN** the ConversationAnalyzer extracts a result with type "pattern"
- **THEN** the result SHALL be saved as a Knowledge entry with category "pattern"
- **AND** the result SHALL additionally be saved as a Learning entry

#### Scenario: Temporal tag preserved
- **WHEN** the ConversationAnalyzer extracts a result with temporal "evergreen"
- **THEN** the Knowledge entry SHALL include tag "temporal:evergreen"

#### Scenario: Session learner uses same routing
- **WHEN** the SessionLearner extracts results
- **THEN** it SHALL use the same all-as-knowledge + dual-save routing as ConversationAnalyzer

#### Scenario: Shared save helper
- **WHEN** ConversationAnalyzer or SessionLearner saves a result
- **THEN** both SHALL delegate to the shared `saveAnalysisResult()` helper with appropriate parameters

### Requirement: Analysis triggers on turn count or token threshold
The system SHALL trigger conversation analysis when either the configured turn threshold (default 10) or token threshold (default 2000) is exceeded since the last analysis for that session.

#### Scenario: Turn threshold triggers analysis
- **WHEN** 10 new turns accumulate since last analysis
- **THEN** analysis SHALL be triggered for that session

#### Scenario: Token threshold triggers analysis
- **WHEN** 2000 tokens of new content accumulate since last analysis (even if fewer than 10 turns)
- **THEN** analysis SHALL be triggered for that session

### Requirement: Session learner extracts high-confidence knowledge at session end
The system SHALL analyze the full session at session end/pause to produce high-confidence knowledge entries.

#### Scenario: Skip short sessions
- **WHEN** a session ends with fewer than 4 turns
- **THEN** no session-end analysis SHALL occur

#### Scenario: Sample long sessions
- **WHEN** a session ends with more than 20 turns
- **THEN** the system SHALL sample first 3 + every 5th + last 5 messages for analysis

#### Scenario: Only store high-confidence results
- **WHEN** session analysis produces results
- **THEN** only entries with confidence == "high" SHALL be stored

### Requirement: Analysis buffer provides async processing
The system SHALL process conversation analysis asynchronously via a buffer with Start/Trigger/TriggerSessionEnd/Stop lifecycle.

#### Scenario: Buffer start and stop
- **WHEN** the application starts and knowledge+observational memory are enabled
- **THEN** the analysis buffer SHALL start a background goroutine and stop cleanly on shutdown

#### Scenario: Queue full drops gracefully
- **WHEN** the analysis queue is full
- **THEN** the request SHALL be dropped with a warn-level log message
