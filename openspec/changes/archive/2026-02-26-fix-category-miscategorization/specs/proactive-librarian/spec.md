## MODIFIED Requirements

### Requirement: Extraction Processing
The proactive librarian extraction pipeline SHALL validate the type of each extraction before saving. When an extraction has an unrecognized type, the system SHALL log a warning and skip that extraction without affecting other extractions in the batch.

#### Scenario: Valid extraction type saved
- **WHEN** an extraction with a recognized type (preference, fact, rule, definition, pattern, correction) meets the auto-save confidence threshold
- **THEN** the knowledge entry SHALL be saved with the correct category

#### Scenario: Unknown extraction type skipped
- **WHEN** an extraction with an unrecognized type is encountered
- **THEN** the system SHALL log a warning with the key and type, skip that extraction, and continue processing remaining extractions

### Requirement: Inquiry Answer Category Validation
The `InquiryProcessor` SHALL validate the category of matched knowledge through `mapCategory()` before saving. Raw casting of LLM-provided category strings to `entknowledge.Category` SHALL NOT be used.

#### Scenario: Valid inquiry answer category
- **WHEN** an inquiry answer match contains a recognized category
- **THEN** the knowledge SHALL be saved and the inquiry resolved

#### Scenario: Invalid inquiry answer category
- **WHEN** an inquiry answer match contains an unrecognized category
- **THEN** the knowledge save SHALL be skipped with a warning log, but the inquiry SHALL still be resolved

### Requirement: Observation Analyzer Prompt Types
The observation analyzer prompt SHALL list all valid extraction types including `pattern` and `correction` in addition to `preference`, `fact`, `rule`, `definition`.

#### Scenario: Prompt includes all types
- **WHEN** the observation analyzer generates its LLM prompt
- **THEN** the type field description SHALL include `preference|fact|rule|definition|pattern|correction`
