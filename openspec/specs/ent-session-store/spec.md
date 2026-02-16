## ADDED Requirements

### Requirement: Ent schema definition
The system SHALL define ent schemas for Session and Message entities with appropriate fields and relationships.

#### Scenario: Session schema generation
- **WHEN** `go generate ./ent` is run
- **THEN** the system SHALL generate type-safe Go code for Session CRUD operations

#### Scenario: Message relationship
- **WHEN** a Session is queried
- **THEN** the system SHALL be able to eagerly load associated Messages

### Requirement: Pure Go SQLite driver
The system SHALL use `modernc.org/sqlite` as the SQLite driver to eliminate CGO dependency.

#### Scenario: Building without CGO
- **WHEN** the project is built with `CGO_ENABLED=0`
- **THEN** the build SHALL succeed without errors

### Requirement: Store interface compatibility
The new ent-based store SHALL implement the existing `Store` interface. The ADK `SessionServiceAdapter` layer wrapping the store SHALL implement get-or-create semantics: when `Get()` returns a "session not found" error, the adapter SHALL call `Create()` to auto-create the session and return it.

#### Scenario: Interface compliance
- **WHEN** `EntStore` is used
- **THEN** it SHALL implement Create, Get, Update, Delete, AppendMessage, and Close methods

#### Scenario: ADK adapter auto-creation on miss
- **WHEN** `SessionServiceAdapter.Get()` receives a "session not found" error from the store
- **THEN** it SHALL call `SessionServiceAdapter.Create()` with the requested session ID and return the newly created session

### Requirement: Automatic schema migration
The system SHALL automatically migrate the database schema on startup.

#### Scenario: New database
- **WHEN** the application starts with a new database file
- **THEN** ent SHALL create the required tables automatically

#### Scenario: Schema update
- **WHEN** the ent schema is modified and app restarts
- **THEN** ent SHALL apply additive migrations without data loss

### Requirement: Observation Entity Schema
The system SHALL define an Ent schema for Observation entities.

#### Scenario: Observation schema fields
- **WHEN** the Observation schema is defined
- **THEN** it SHALL include: id (UUID), session_key (string, not empty), content (text, not empty), token_count (int, default 0), source_start_index (int), source_end_index (int), created_at (time, default now)

#### Scenario: Observation schema indexes
- **WHEN** querying observations
- **THEN** the schema SHALL have indexes on session_key and created_at for efficient retrieval

#### Scenario: Observation auto-migration
- **WHEN** the application starts
- **THEN** ent SHALL create the Observation table automatically if it does not exist

### Requirement: Reflection Entity Schema
The system SHALL define an Ent schema for Reflection entities.

#### Scenario: Reflection schema fields
- **WHEN** the Reflection schema is defined
- **THEN** it SHALL include: id (UUID), session_key (string, not empty), content (text, not empty), token_count (int, default 0), generation (int, default 1), created_at (time, default now)

#### Scenario: Reflection schema indexes
- **WHEN** querying reflections
- **THEN** the schema SHALL have indexes on session_key and created_at for efficient retrieval

#### Scenario: Reflection auto-migration
- **WHEN** the application starts
- **THEN** ent SHALL create the Reflection table automatically if it does not exist

### Requirement: OM Data Access Methods
The system SHALL provide data access methods for observations and reflections through the session store.

#### Scenario: Save observation
- **WHEN** an observation is generated
- **THEN** the store SHALL persist the observation with all fields

#### Scenario: List observations by session
- **WHEN** querying observations for a session
- **THEN** the store SHALL return observations ordered by created_at ascending

#### Scenario: Delete observations by session
- **WHEN** observations are condensed into a reflection
- **THEN** the store SHALL delete the specified observations

#### Scenario: Save reflection
- **WHEN** a reflection is generated
- **THEN** the store SHALL persist the reflection with all fields

#### Scenario: List reflections by session
- **WHEN** querying reflections for a session
- **THEN** the store SHALL return reflections ordered by created_at ascending

### Requirement: Message Author Field in Ent Schema
The Message ent schema SHALL include an optional `author` string field with a default empty value, used to persist the ADK agent name for multi-agent routing.

#### Scenario: New message with author
- **WHEN** a message is saved with a non-empty Author
- **THEN** the `author` column SHALL store the agent name

#### Scenario: Legacy message without author
- **WHEN** an existing message has no author value
- **THEN** the `author` column SHALL default to empty string
