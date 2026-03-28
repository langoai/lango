## ADDED Requirements

### Requirement: Streaming cursor blink animation
During agent streaming (stateStreaming), the chat view SHALL display a blinking block cursor ("▌") appended to the stream content. The cursor SHALL toggle visibility every 400ms via tea.Tick.

#### Scenario: Cursor appears during streaming
- **WHEN** the first ChunkMsg arrives during stateStreaming
- **THEN** a cursor blink tick SHALL start and "▌" SHALL appear after stream content

#### Scenario: Cursor toggles on tick
- **WHEN** CursorTickMsg fires during stateStreaming
- **THEN** showCursor SHALL toggle and the next tick SHALL be scheduled

#### Scenario: Cursor stops after streaming ends
- **WHEN** DoneMsg or ErrorMsg is received
- **THEN** showCursor SHALL be false and no further ticks SHALL be scheduled

### Requirement: Tick dedup guard
Multiple ChunkMsg arrivals SHALL NOT create duplicate tick timers. A cursorTickActive flag SHALL prevent redundant tick creation.

#### Scenario: Rapid chunks don't multiply ticks
- **WHEN** multiple ChunkMsg arrive while cursorTickActive is true
- **THEN** no additional tick commands SHALL be created
