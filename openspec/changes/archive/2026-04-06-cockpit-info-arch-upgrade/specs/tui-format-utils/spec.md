## ADDED Requirements

### Requirement: Shared string truncation
The `tui` package SHALL export a `Truncate(s string, maxLen int) string` function that shortens strings exceeding `maxLen` by appending "..." and trimming to fit.

#### Scenario: String within limit
- **WHEN** `Truncate("hello", 10)` is called
- **THEN** the result SHALL be `"hello"` unchanged

#### Scenario: String exceeds limit
- **WHEN** `Truncate("hello world", 8)` is called
- **THEN** the result SHALL be `"hello..."` (truncated to 8 characters including ellipsis)

### Requirement: Shared word wrap
The `tui` package SHALL export a `WordWrap(text string, width int) string` function that wraps text at word boundaries to fit within the given width.

#### Scenario: Text within width
- **WHEN** `WordWrap("short", 80)` is called
- **THEN** the result SHALL be `"short"` unchanged

#### Scenario: Text exceeding width
- **WHEN** text exceeds the specified width
- **THEN** line breaks SHALL be inserted at word boundaries

### Requirement: Shared number formatting
The `tui` package SHALL export `FormatNumber(n int64) string` and `FormatTokens(n int) string` functions that render integers with comma-separated thousands.

#### Scenario: FormatNumber with thousands
- **WHEN** `FormatNumber(12345)` is called
- **THEN** the result SHALL be `"12,345"`

#### Scenario: FormatTokens delegates to FormatNumber
- **WHEN** `FormatTokens(12345)` is called
- **THEN** the result SHALL be `"12,345"` (same as `FormatNumber(int64(12345))`)

### Requirement: Shared duration formatting
The `tui` package SHALL export a `FormatDuration(d time.Duration) string` function that renders durations in human-readable form.

#### Scenario: Duration in hours and minutes
- **WHEN** `FormatDuration(2*time.Hour + 15*time.Minute)` is called
- **THEN** the result SHALL be `"2h 15m"`

#### Scenario: Sub-second duration
- **WHEN** `FormatDuration(150 * time.Millisecond)` is called
- **THEN** the result SHALL be `"150ms"`

### Requirement: Shared relative time formatting with two variants
The `tui` package SHALL export `RelativeTime(now, t time.Time) string` (precise) and `RelativeTimeHuman(now, t time.Time) string` (friendly) functions.

#### Scenario: RelativeTime sub-minute precision
- **WHEN** `RelativeTime(now, now.Add(-5*time.Second))` is called
- **THEN** the result SHALL be `"5s ago"`

#### Scenario: RelativeTimeHuman sub-minute friendly
- **WHEN** `RelativeTimeHuman(now, now.Add(-5*time.Second))` is called
- **THEN** the result SHALL be `"just now"`

#### Scenario: Both variants identical for minutes and above
- **WHEN** elapsed time is 1 minute or more
- **THEN** both `RelativeTime` and `RelativeTimeHuman` SHALL return the same format (e.g., `"5m ago"`, `"2h ago"`)
