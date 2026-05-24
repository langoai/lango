## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: Stored status transcript content is normalized to single-line text
- **WHEN** the chat model appends a compact `status` transcript item whose content contains embedded newlines or terminal control sequences
- **THEN** the stored status content SHALL already be stripped and normalized to single-line text before rendering
