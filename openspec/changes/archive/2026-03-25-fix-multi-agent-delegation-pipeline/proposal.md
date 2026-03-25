## Why

Multi-agent structured orchestration mode에서 vault 핸드오프가 3가지 별개의 버그로 실패.
(1) ADK 2-phase delegation과 Lango 가드 타이밍 충돌, (2) 병렬 tool response의 ID 소실,
(3) 보조 이슈(TUI 로그 누수, dangling cleanup author 손실, recovery 무한 루프).
DB/turntrace 증거로 각 원인이 확정됨.

## What Changes

- `transfer_to_agent` FunctionCall이 orchestrator direct-tool guard를 통과하도록 예외 추가 (ADK 2-phase 호환)
- `convertMessages`에서 EventsAdapter가 merge한 다수 FunctionResponse를 개별 provider.Message로 분리
- `closeDanglingParentToolCalls`의 합성 tool response에 originating agent Author 보존
- `CauseOrchestratorDirectTool` recovery를 Escalate로 변경 (same-input retry 무한 루프 방지)
- TUI에서 Go stdlib logger를 로그 파일로 리다이렉트 (ADK의 log.Printf 누수 차단)
- `repairOrphanedToolCalls` 에러 메시지를 더 정확하고 actionable하게 개선
- CoordinatingExecutor recovery에 error classification 진단 로깅 추가

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `agent-control-plane`: orchestrator direct-tool guard에 transfer_to_agent 예외 추가, recovery에 CauseOrchestratorDirectTool escalation 추가
- `agent-error-handling`: repairOrphanedToolCalls 메시지 개선, dangling cleanup author 보존, convertMessages FunctionResponse split

## Impact

- `internal/adk/agent.go` — isPureTransferToAgentCall guard exception
- `internal/adk/model.go` — convertMessages multi-FunctionResponse split
- `internal/adk/session_service.go` — danglingCall OriginAuthor, diagnostic logging
- `internal/agentrt/coordinating_executor.go` — recovery diagnostic logging
- `internal/agentrt/recovery.go` — CauseOrchestratorDirectTool escalation
- `internal/provider/openai/openai.go` — improved orphan error message
- `cmd/lango/main.go` — stdlib logger redirect
