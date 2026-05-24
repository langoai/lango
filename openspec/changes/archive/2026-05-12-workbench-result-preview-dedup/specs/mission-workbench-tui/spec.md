## ADDED Requirements

### Requirement: Completed-turn result preview avoids redundant assistant labels

The standalone workbench SHALL avoid repeating the assistant label inside the completed-turn body preview when the preview already has a `Last result:` prefix.

#### Scenario: Success preview drops assistant label duplication
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the latest completed-turn summary is an assistant success summary
- **THEN** the body SHALL show `Last result: <summary>` without repeating `Assistant reply:`
