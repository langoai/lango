## MODIFIED Requirements

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
