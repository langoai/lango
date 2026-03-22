## Why

Recent refactoring commits moved many agent tool builders out of `internal/app/`, introduced package-owned agent memory tools, added builder parity coverage, and tightened agent memory validation. The main OpenSpec specs still describe only part of that structure, leaving the documented contract behind the code.

## What Changes

- Document package-owned tool builders for automation, data, collaboration, sentinel, and foundation tool packages.
- Update the agent memory spec to reflect the `Entry`/`InMemoryStore` names used in code, package-owned tool registration, context-aware recall with kind filtering, and runtime kind validation.
- Document shared automation helper interfaces and channel-detection helper used by cron, background, and workflow packages.
- Update tool catalog expectations for sentinel tool registration and the remaining app-owned builder exceptions.
- Add parity-verification coverage for extracted tool builders so stable tool names, non-nil handlers, and duplicate-name safety are specified.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `agent-memory`: align names and behavior with the current package, including tool ownership and kind-aware context fallback.
- `automation-agent-tools`: document shared automation runner/sender interfaces and context-based delivery detection.
- `domain-tool-builders`: expand package-owned builder requirements beyond economy.
- `tool-catalog`: update sentinel registration and app-owned builder exceptions.
- `parity-verification`: add extracted tool builder parity coverage.

## Impact

- Affected code: `internal/agentmemory/`, `internal/automation/`, `internal/cron/`, `internal/background/`, `internal/workflow/`, `internal/graph/`, `internal/embedding/`, `internal/librarian/`, `internal/memory/`, `internal/p2p/team/`, `internal/economy/escrow/sentinel/`, `internal/tooloutput/`, `internal/tools/{browser,exec,filesystem,crypto,secrets}/`, and `internal/app/modules.go`.
- Affected specs: `agent-memory`, `automation-agent-tools`, `domain-tool-builders`, `tool-catalog`, `parity-verification`.
- Verification sources: recent commits `4f59b8c`, `eddbe4b`, `c3a9322`, and `bd97b6c`.
