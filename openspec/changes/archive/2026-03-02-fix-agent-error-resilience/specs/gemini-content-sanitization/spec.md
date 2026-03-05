## ADDED Requirements

### Requirement: Gemini content turn-order sanitization pipeline
The Gemini provider SHALL sanitize the content sequence before every API call to satisfy Gemini's strict turn-ordering rules. The sanitization pipeline SHALL execute 5 steps in order: (1) drop leading orphaned FunctionResponses, (2) merge consecutive same-role contents, (3) prepend synthetic user turn if sequence starts with model, (4) ensure FunctionCall/FunctionResponse pairing, (5) final merge pass.

#### Scenario: Consecutive same-role contents merged
- **WHEN** the content sequence contains consecutive entries with the same role (e.g., model, model)
- **THEN** the sanitizer SHALL merge their Parts into a single Content entry with that role

#### Scenario: Sequence starting with model turn
- **WHEN** the first content entry has role "model"
- **THEN** the sanitizer SHALL prepend a synthetic user Content with text "[continue]"

#### Scenario: Orphaned FunctionResponse at start of sequence
- **WHEN** the content sequence starts with user-role entries containing only FunctionResponse parts (no preceding FunctionCall)
- **THEN** the sanitizer SHALL drop those entries

#### Scenario: FunctionCall without matching FunctionResponse
- **WHEN** a model Content contains FunctionCall parts and the next Content is not a user FunctionResponse
- **THEN** the sanitizer SHALL insert a synthetic user Content with FunctionResponse parts (status "[no response available]") for each FunctionCall

#### Scenario: Valid FunctionCall/FunctionResponse pair preserved
- **WHEN** a model Content with FunctionCall is immediately followed by a user Content with matching FunctionResponse
- **THEN** the sanitizer SHALL pass the pair through unchanged

#### Scenario: Empty content sequence
- **WHEN** the content sequence is empty
- **THEN** the sanitizer SHALL return the empty sequence unchanged

### Requirement: Consecutive role merging in session events
The EventsAdapter.All() method SHALL merge consecutive same-role events as a defense-in-depth measure to prevent turn-order violations at the ADK/provider boundary. Parts from consecutive same-role events SHALL be concatenated into a single event.

#### Scenario: Consecutive assistant events merged
- **WHEN** the event history contains two consecutive events with role "model"
- **THEN** EventsAdapter.All() SHALL yield a single event with both events' Parts concatenated

#### Scenario: Alternating roles preserved
- **WHEN** the event history contains alternating user and model events
- **THEN** EventsAdapter.All() SHALL yield each event separately without merging

#### Scenario: Len() consistent with All()
- **WHEN** consecutive same-role events exist in history
- **THEN** EventsAdapter.Len() SHALL return the count of merged events (matching All() output), not the raw history count
