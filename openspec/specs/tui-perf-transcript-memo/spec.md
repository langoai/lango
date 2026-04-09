## Purpose

Capability spec for tui-perf-transcript-memo. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Transcript block memoization
The chatViewModel SHALL cache rendered block output per transcriptItem in a `cachedBlock` field. The `render()` method SHALL skip re-rendering entries with a non-empty cachedBlock and reuse the cached string.

#### Scenario: Cached entries skip rendering
- **WHEN** render() iterates entries where cachedBlock is non-empty
- **THEN** the cached string SHALL be used directly without calling renderEntry()

#### Scenario: New entries get cached on first render
- **WHEN** render() encounters an entry with empty cachedBlock
- **THEN** renderEntry() SHALL be called and the result stored in cachedBlock

#### Scenario: Width change invalidates all caches
- **WHEN** setSize() detects a width change
- **THEN** all entries' cachedBlock SHALL be cleared to empty string

#### Scenario: Tool finalization invalidates specific cache
- **WHEN** finalizeToolResult() updates a tool entry's state and output
- **THEN** that entry's cachedBlock SHALL be cleared

#### Scenario: Thinking finalization invalidates specific cache
- **WHEN** finalizeThinking() updates a thinking entry's summary and duration
- **THEN** that entry's cachedBlock SHALL be cleared

### Requirement: Transcript entry trimming with cap
The chatViewModel SHALL enforce a maximum of 2000 transcript entries. When the cap is exceeded, the oldest entries SHALL be trimmed with a tombstone summary.

#### Scenario: Entries trimmed when cap exceeded
- **WHEN** appendEntry() causes entries to exceed maxTranscriptEntries (2000)
- **THEN** the oldest maxTranscriptEntries/4 entries SHALL be removed and a tombstone entry inserted

#### Scenario: Tombstone count accumulates across trims
- **WHEN** trimming occurs and a tombstone already exists at entries[0]
- **THEN** the new tombstone SHALL show the accumulated total of all trimmed entries

#### Scenario: In-flight tool/thinking entries preserved during trim
- **WHEN** active tool (state=running) or thinking (state=active) entries exist in the trim range
- **THEN** those entries SHALL be collected and preserved after the tombstone

#### Scenario: New backing array releases old memory
- **WHEN** trimming occurs
- **THEN** a new slice SHALL be allocated via make() so the old backing array is eligible for GC
