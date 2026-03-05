# Tasks

## 1. Configuration
- [x] 1.1 Create `internal/config/types_mcp.go` with MCPConfig and MCPServerConfig types
- [x] 1.2 Add MCP field to Config struct in `internal/config/types.go`
- [x] 1.3 Add MCP defaults to `DefaultConfig()` in `internal/config/loader.go`
- [x] 1.4 Add MCP viper defaults in `Load()`
- [x] 1.5 Add MCP env var substitution in `substituteEnvVars()`
- [x] 1.6 Add MCP validation in `Validate()`

## 2. MCP Core Package
- [x] 2.1 Create `internal/mcp/errors.go` with sentinel errors
- [x] 2.2 Create `internal/mcp/env.go` with ExpandEnv, ExpandEnvMap, BuildEnvSlice
- [x] 2.3 Create `internal/mcp/env_test.go` with tests
- [x] 2.4 Create `internal/mcp/connection.go` with ServerConnection (transport, connect, disconnect, health check, reconnect)
- [x] 2.5 Create `internal/mcp/manager.go` with ServerManager (ConnectAll, DisconnectAll, AllTools, ServerStatus)
- [x] 2.6 Create `internal/mcp/adapter.go` with AdaptTools/AdaptTool (naming, schema mapping, handler proxy, truncation)
- [x] 2.7 Create `internal/mcp/adapter_test.go` with tests for buildParams, parseSafetyLevel
- [x] 2.8 Create `internal/mcp/config_loader.go` with MergedServers, LoadMCPFile, SaveMCPFile

## 3. App Integration
- [x] 3.1 Create `internal/app/wiring_mcp.go` with initMCP() and buildMCPManagementTools()
- [x] 3.2 Add MCPManager field to App struct in `internal/app/types.go`
- [x] 3.3 Wire MCP into init sequence in `internal/app/app.go` (step 5n, after dispatcher tools)
- [x] 3.4 Register MCP lifecycle component (PriorityNetwork) for graceful shutdown
- [x] 3.5 Register MCP auth headers with secret scanner
- [x] 3.6 Add `lango mcp` to blockLangoExec guard in `internal/app/tools.go`

## 4. CLI Commands
- [x] 4.1 Create `internal/cli/mcp/mcp.go` root command
- [x] 4.2 Create `internal/cli/mcp/list.go` — list configured servers
- [x] 4.3 Create `internal/cli/mcp/add.go` — add server with transport/command/url/env/headers/scope
- [x] 4.4 Create `internal/cli/mcp/remove.go` — remove server
- [x] 4.5 Create `internal/cli/mcp/get.go` — show server details + discovered tools
- [x] 4.6 Create `internal/cli/mcp/test.go` — test connectivity (handshake, tools, ping)
- [x] 4.7 Create `internal/cli/mcp/enable.go` — enable/disable server toggle
- [x] 4.8 Register `lango mcp` command in `cmd/lango/main.go` (GroupID: "infra")

## 5. Dependencies
- [x] 5.1 Add `github.com/modelcontextprotocol/go-sdk` v1.4.0
- [x] 5.2 Run `go mod tidy`

## 6. Verification
- [x] 6.1 `go build ./...` passes
- [x] 6.2 `go test ./internal/mcp/...` passes
- [x] 6.3 `go test ./internal/config/...` passes
- [x] 6.4 `go test ./...` full suite passes
