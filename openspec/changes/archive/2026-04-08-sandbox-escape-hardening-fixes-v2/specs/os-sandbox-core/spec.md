## ADDED Requirements

### Requirement: bwrap mount ordering
`compileBwrapArgs` SHALL emit the root-level mount (`--ro-bind / /` when `ReadOnlyGlobal=true`, or `--ro-bind <p> <p>` entries for explicit `ReadPaths`) BEFORE the specialised mounts `--proc /proc`, `--dev /dev`, and `--tmpfs /run`. bubblewrap processes options left-to-right, and a later root bind would shadow any earlier mounts nested under the sandbox root, leaking the host's `/proc` and `/dev` into the sandboxed child and weakening PID namespace and device isolation. The specialised mounts must therefore be layered on top of the root bind, not underneath it.

#### Scenario: Root bind precedes --proc
- **WHEN** `compileBwrapArgs` is called with `ReadOnlyGlobal=true`
- **THEN** the index of `--ro-bind / /` in the returned argv slice SHALL be less than the index of `--proc /proc`
- **AND** the index of `--ro-bind / /` SHALL be less than the indices of `--dev /dev` and `--tmpfs /run`

#### Scenario: Specialised mounts still present
- **WHEN** `compileBwrapArgs` is called with any valid Policy
- **THEN** the returned argv slice SHALL contain `--proc /proc`, `--dev /dev`, and `--tmpfs /run` as three-token pairs, unconditionally

### Requirement: Sandbox path validation against DataRoot overlap
`config.Validate` SHALL reject configurations where `sandbox.workspacePath` or any entry of `sandbox.allowedWritePaths` resolves to `cfg.DataRoot` itself or to a subtree of `cfg.DataRoot`. This check is necessary because `DefaultToolPolicy` adds `cfg.DataRoot` to `DenyPaths`, and the resulting `--tmpfs cfg.DataRoot` mount (bwrap) or `(deny file* (subpath ...))` rule (Seatbelt) would cover the workspace and make it silently unreachable at runtime. The validation check SHALL fire AFTER `NormalizePaths` so it catches both relative paths that were resolved under `DataRoot` and absolute paths the user explicitly wrote inside the control-plane.

The validation error message SHALL name the colliding path, state that it is inside `cfg.DataRoot`, and direct the user to use an absolute path outside the control-plane.

#### Scenario: workspacePath nested under DataRoot rejected
- **WHEN** `cfg.DataRoot = "/tmp/lango"` and `cfg.Sandbox.WorkspacePath = "/tmp/lango/repo"`
- **THEN** `Validate(cfg)` SHALL return an error
- **AND** the error message SHALL contain `"sandbox.workspacePath"` and `"inside cfg.DataRoot"`

#### Scenario: workspacePath equal to DataRoot rejected
- **WHEN** `cfg.DataRoot = "/tmp/lango"` and `cfg.Sandbox.WorkspacePath = "/tmp/lango"`
- **THEN** `Validate(cfg)` SHALL return an error mentioning `sandbox.workspacePath`

#### Scenario: workspacePath outside DataRoot accepted
- **WHEN** `cfg.DataRoot = "/tmp/lango"` and `cfg.Sandbox.WorkspacePath = "/tmp/some-other-dir"`
- **THEN** `Validate(cfg)` SHALL return nil

#### Scenario: allowedWritePaths entry nested under DataRoot rejected
- **WHEN** `cfg.DataRoot = "/tmp/lango"` and `cfg.Sandbox.AllowedWritePaths` contains `"/tmp/lango/scratch"`
- **THEN** `Validate(cfg)` SHALL return an error naming `sandbox.allowedWritePaths` and the offending entry

#### Scenario: Empty workspacePath accepted
- **WHEN** `cfg.Sandbox.WorkspacePath = ""`
- **THEN** `Validate(cfg)` SHALL NOT error on the workspace path check (the supervisor falls back to `os.Getwd()` at runtime)
