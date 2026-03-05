## 1. README.md Updates

- [x] 1.1 Add MCP Integration to Features list (after Workflow Engine, before Secure)
- [x] 1.2 Add `lango mcp` CLI commands block to CLI Commands section (after workflow, before p2p)
- [x] 1.3 Add `mcp/` to Architecture diagram cli/ tree (after workflow/)
- [x] 1.4 Add `mcp/` to Architecture diagram internal/ tree (before toolcatalog/)

## 2. docs/cli/index.md Update

- [x] 2.1 Add MCP Servers section to Quick Reference table (after Automation section)

## 3. docs/cli/mcp.md Creation

- [x] 3.1 Create docs/cli/mcp.md with full MCP CLI reference (all 7 subcommands with flag tables and examples)
- [x] 3.2 Include Configuration section documenting 3-scope merge, TUI settings, key config options, and tool naming convention

## 4. Verification

- [x] 4.1 Run `go build ./...` to confirm no code regressions
- [x] 4.2 Run `go run ./cmd/lango mcp --help` to verify CLI commands match documentation
