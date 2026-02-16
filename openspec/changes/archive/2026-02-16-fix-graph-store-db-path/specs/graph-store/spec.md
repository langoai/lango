## MODIFIED Requirements

### Requirement: BoltDB Store Initialization
`NewBoltStore(path)` SHALL expand tilde prefixes and ensure parent directories exist before opening the BoltDB database.

The function SHALL:
1. If `path` starts with `~/`, replace the prefix with the user's home directory obtained from `os.UserHomeDir()`
2. Create parent directories using `os.MkdirAll(filepath.Dir(path), 0o700)` if they do not exist
3. Open the BoltDB file at the resolved path

#### Scenario: Path with tilde prefix
- **WHEN** `NewBoltStore("~/.lango/graph.db")` is called
- **THEN** the tilde SHALL be expanded to the user's home directory and the database SHALL be opened at the absolute path

#### Scenario: Parent directory does not exist
- **WHEN** `NewBoltStore` is called with a path whose parent directory does not exist
- **THEN** the parent directory SHALL be created with `0o700` permissions before opening the database

#### Scenario: Home directory resolution failure
- **WHEN** `os.UserHomeDir()` returns an error during tilde expansion
- **THEN** `NewBoltStore` SHALL return a wrapped error without attempting to open the database

#### Scenario: Directory creation failure
- **WHEN** `os.MkdirAll` returns an error
- **THEN** `NewBoltStore` SHALL return a wrapped error without attempting to open the database
