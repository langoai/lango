## 1. Graph Store API Extension

- [x] 1.1 Add Count, PredicateStats, ClearAll to Store interface in internal/graph/store.go
- [x] 1.2 Implement Count on BoltStore using bucket Stats().KeyN
- [x] 1.3 Implement PredicateStats on BoltStore by iterating SPO bucket and splitting keys
- [x] 1.4 Implement ClearAll on BoltStore by deleting and recreating all three buckets
- [x] 1.5 Update fakeGraphStore in internal/learning/graph_engine_test.go with new methods

## 2. Graph CLI Commands

- [x] 2.1 Create internal/cli/graph/graph.go with NewGraphCmd and initGraphStore helper
- [x] 2.2 Create internal/cli/graph/status.go — graph status [--json]
- [x] 2.3 Create internal/cli/graph/query.go — graph query --subject/--predicate/--object [--json] [--limit]
- [x] 2.4 Create internal/cli/graph/stats.go — graph stats [--json]
- [x] 2.5 Create internal/cli/graph/clear.go — graph clear [--force]
- [x] 2.6 Wire cligraph.NewGraphCmd in cmd/lango/main.go

## 3. Agent CLI Commands

- [x] 3.1 Create internal/cli/agent/agent.go with NewAgentCmd
- [x] 3.2 Create internal/cli/agent/status.go — agent status [--json]
- [x] 3.3 Create internal/cli/agent/list.go — agent list [--json] [--check]
- [x] 3.4 Wire cliagent.NewAgentCmd in cmd/lango/main.go

## 4. Doctor Checks

- [x] 4.1 Create internal/cli/doctor/checks/graph_store.go — GraphStoreCheck
- [x] 4.2 Create internal/cli/doctor/checks/multi_agent.go — MultiAgentCheck
- [x] 4.3 Create internal/cli/doctor/checks/a2a.go — A2ACheck
- [x] 4.4 Register all three checks in AllChecks() in checks.go

## 5. Onboard Wizard Screens

- [x] 5.1 Add graph, multi_agent, a2a menu items in menu.go (before save)
- [x] 5.2 Create NewGraphForm in forms_impl.go
- [x] 5.3 Create NewMultiAgentForm in forms_impl.go
- [x] 5.4 Create NewA2AForm in forms_impl.go
- [x] 5.5 Add menu routing cases in wizard.go handleMenuSelection
- [x] 5.6 Add config write-back cases in state_update.go UpdateConfigFromForm

## 6. Config Defaults and Validation

- [x] 6.1 Add Graph defaults to DefaultConfig in loader.go
- [x] 6.2 Add A2A defaults to DefaultConfig in loader.go
- [x] 6.3 Add Viper defaults for graph and a2a fields in Load
- [x] 6.4 Add graph.backend validation in Validate
- [x] 6.5 Add a2a.baseUrl and a2a.agentName validation in Validate

## 7. README Documentation

- [x] 7.1 Add Knowledge Graph, Multi-Agent, A2A feature bullets
- [x] 7.2 Add onboard wizard items 7-9
- [x] 7.3 Add graph and agent CLI commands to CLI section
- [x] 7.4 Add architecture tree entries for new packages
- [x] 7.5 Add configuration reference rows for graph, multi-agent, a2a
- [x] 7.6 Add Knowledge Graph & Graph RAG section
- [x] 7.7 Add Multi-Agent Orchestration section
- [x] 7.8 Add A2A Protocol section

## 8. Verification

- [x] 8.1 go build ./... passes
- [x] 8.2 go test ./... passes
- [x] 8.3 go vet ./... passes
