## MODIFIED Requirements

### Requirement: ListAgentNames method on Store interface
The agentmemory.Store interface SHALL include a `ListAgentNames() ([]string, error)` method that returns the names of all agents that have stored memories. This method is required to support the `memory agents` CLI command.

#### Scenario: ListAgentNames with entries
- **WHEN** ListAgentNames() is called on a store containing entries for agents "researcher" and "planner"
- **THEN** the method returns ["researcher", "planner"] (order not guaranteed) with no error

#### Scenario: ListAgentNames with no entries
- **WHEN** ListAgentNames() is called on a store with no agent memory entries
- **THEN** the method returns an empty slice with no error

### Requirement: ListAll method on Store interface
The agentmemory.Store interface SHALL include a `ListAll(agentName string) ([]*Entry, error)` method that returns all memory entries for the specified agent. This method is required to support the `memory agent <name>` CLI command.

#### Scenario: ListAll for existing agent
- **WHEN** ListAll("researcher") is called on a store containing 5 entries for "researcher"
- **THEN** the method returns all 5 Entry pointers with no error

#### Scenario: ListAll for nonexistent agent
- **WHEN** ListAll("unknown") is called on a store with no entries for "unknown"
- **THEN** the method returns an empty slice with no error

### Requirement: MemStore implementation
The in-memory agentmemory.MemStore implementation SHALL implement both ListAgentNames() and ListAll() by iterating the internal memory map.

#### Scenario: MemStore ListAgentNames
- **WHEN** ListAgentNames() is called on a MemStore with entries for 3 agents
- **THEN** the method returns a slice of 3 agent name strings

### Requirement: Backward compatibility
The addition of ListAgentNames() and ListAll() to the Store interface SHALL NOT change the behavior of existing Store methods. All existing tests SHALL continue to pass.

#### Scenario: Existing tests pass
- **WHEN** `go test ./internal/agentmemory/...` is run after the interface additions
- **THEN** all existing tests pass without modification
